package crdinstaller

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apimachineryYaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"

	"github.com/deckhouse/module-sdk/pkg/crd-installer/openapi"
	"github.com/deckhouse/module-sdk/pkg/utils"
)

const (
	LabelHeritage string = "heritage"
)

// 1Mb - maximum size of kubernetes object
// if we take less, we have to handle io.ErrShortBuffer error and increase the buffer
// take more does not make any sense due to kubernetes limitations
// Considering that etcd has a default value of 1.5Mb, it was decided to set it to 2Mb,
// so that in most cases we would get a more informative error from Kubernetes, not just "short buffer"
const bufSize = 2 * 1024 * 1024

var crdGVR = schema.GroupVersionResource{
	Group:    "apiextensions.k8s.io",
	Version:  "v1",
	Resource: "customresourcedefinitions",
}

func WithExtraLabels(labels map[string]string) InstallerOption {
	return func(installer *CRDsInstaller) {
		installer.crdExtraLabels = labels
	}
}

func WithFileFilter(fn func(path string) bool) InstallerOption {
	return func(installer *CRDsInstaller) {
		installer.fileFilter = fn
	}
}

// CRDsInstaller simultaneously installs CRDs from specified directory
type CRDsInstaller struct {
	k8sClient     dynamic.Interface
	crdFilesPaths []string
	buffer        []byte

	// concurrent tasks to create resource in a k8s cluster
	k8sTasks *multierror.Group

	crdExtraLabels map[string]string
	fileFilter     func(path string) bool

	appliedGVKsLock sync.Mutex

	// list of GVKs, applied to the cluster
	appliedGVKs []string
}

func (cp *CRDsInstaller) GetAppliedGVKs() []string {
	return cp.appliedGVKs
}

func (cp *CRDsInstaller) Run(ctx context.Context) error {
	var errs error

	for _, crdFilePath := range cp.crdFilesPaths {
		if cp.fileFilter != nil && !cp.fileFilter(crdFilePath) {
			continue
		}

		err := cp.processCRD(ctx, crdFilePath)
		if err != nil {
			err = fmt.Errorf("error occurred during processing %q file: %w", crdFilePath, err)

			errs = errors.Join(errs, err)

			continue
		}
	}

	terr := cp.k8sTasks.Wait()
	if terr.ErrorOrNil() != nil {
		for _, e := range terr.Errors {
			errs = errors.Join(errs, e)
		}
	}

	return errs
}

func (cp *CRDsInstaller) DeleteCRDs(ctx context.Context, crdsToDelete []string) ([]string, error) {
	var deletedCRDs []string
	// delete crds listed in crdsToDelete if there are no related custom resources in the cluster
	for _, crdName := range crdsToDelete {
		deleteCRD := true

		crd, err := cp.GetCRDFromCluster(ctx, crdName)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return nil, fmt.Errorf("error occurred during %s CRD clean up: %w", crdName, err)
			}

			continue
		}

		for _, version := range crd.Spec.Versions {
			if !version.Storage {
				continue
			}

			gvr := schema.GroupVersionResource{
				Group:    crd.Spec.Group,
				Version:  version.Name,
				Resource: crd.Spec.Names.Plural,
			}

			list, err := cp.k8sClient.Resource(gvr).List(ctx, apimachineryv1.ListOptions{})
			if err != nil {
				return nil, fmt.Errorf("error occurred listing %s CRD objects of version %s: %w", crdName, version.Name, err)
			}

			if len(list.Items) > 0 {
				deleteCRD = false

				break
			}
		}

		if deleteCRD {
			err := cp.k8sClient.Resource(crdGVR).Delete(ctx, crdName, apimachineryv1.DeleteOptions{})
			if err != nil {
				return nil, fmt.Errorf("error occurred deleting %s CRD: %w", crdName, err)
			}

			deletedCRDs = append(deletedCRDs, crdName)
		}
	}

	return deletedCRDs, nil
}

func (cp *CRDsInstaller) processCRD(ctx context.Context, crdFilePath string) error {
	crdFileReader, err := os.Open(crdFilePath)
	if err != nil {
		return err
	}
	defer crdFileReader.Close()

	crdReader := apimachineryYaml.NewDocumentDecoder(crdFileReader)

	for {
		n, err := crdReader.Read(cp.buffer)
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		data := cp.buffer[:n]
		if len(data) == 0 {
			// some empty yaml document, or empty string before separator
			continue
		}

		rd := bytes.NewReader(data)

		err = cp.putCRDToCluster(ctx, rd, n)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cp *CRDsInstaller) putCRDToCluster(ctx context.Context, crdReader io.Reader, bufferSize int) error {
	// Decode into unstructured first: the typed struct below cannot hold
	// x-kubernetes-sensitive-data, and sanitize needs the original document to recover it.
	desired := &unstructured.Unstructured{}
	err := apimachineryYaml.NewYAMLOrJSONDecoder(crdReader, bufferSize).Decode(&desired)
	if err != nil {
		return err
	}

	// a comment or other non-object yaml document decodes to a nil pointer / empty object, skip it
	if desired == nil || len(desired.Object) == 0 {
		return nil
	}

	crd := &apiextensionsv1.CustomResourceDefinition{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(desired.Object, crd); err != nil {
		return err
	}
	if crd.APIVersion != apiextensionsv1.SchemeGroupVersion.String() || crd.Kind != "CustomResourceDefinition" {
		return fmt.Errorf("invalid CRD document apiversion/kind: '%s/%s'", crd.APIVersion, crd.Kind)
	}

	desired, err = sanitize(crd, desired.Object)
	if err != nil {
		return fmt.Errorf("sanitize %s: %w", crd.Name, err)
	}

	cp.k8sTasks.Go(func() error {
		err := cp.updateOrInsertCRD(ctx, crd, desired)
		if err == nil {
			var crdGroup, crdKind string

			if len(crd.Spec.Group) == 0 {
				return fmt.Errorf("process %s: couldn't find CRD's .group key", crd.Name)
			}

			crdGroup = crd.Spec.Group

			if len(crd.Spec.Names.Kind) == 0 {
				return fmt.Errorf("process %s: couldn't find CRD's .spec.names.kind key", crd.Name)
			}

			crdKind = crd.Spec.Names.Kind

			if len(crd.Spec.Versions) == 0 {
				return fmt.Errorf("process %s: couldn't find CRD's .spec.versions key", crd.Name)
			}

			crdVersions := make([]string, 0, len(crd.Spec.Versions))
			for _, version := range crd.Spec.Versions {
				crdVersions = append(crdVersions, version.Name)
			}

			cp.appliedGVKsLock.Lock()

			for _, crdVersion := range crdVersions {
				cp.appliedGVKs = append(cp.appliedGVKs, fmt.Sprintf("%s/%s/%s", crdGroup, crdVersion, crdKind))
			}

			cp.appliedGVKsLock.Unlock()
		}

		return err
	})

	return nil
}

// sanitize reduces the CRD document to the fields the apiserver actually knows.
//
// Everything else — x-doc-examples and friends, plus keys that look official but are
// not, like x-kubernetes-immutable — is dropped here instead of being sent. The
// apiserver prunes them anyway, after logging one "unknown field" warning per
// occurrence, and because it prunes them the stored spec could never equal the desired
// one, so every run issued a pointless Update.
//
// The typed CRD is that pruned document already, at every level. The single exception
// is x-kubernetes-sensitive-data, which apiextensionsv1 does not model but the Deckhouse
// apiserver does, so each version schema is re-decoded through the fork in
// pkg/crd-installer/openapi that carries it.
func sanitize(crd *apiextensionsv1.CustomResourceDefinition, raw map[string]any) (*unstructured.Unstructured, error) {
	clean, err := utils.ToUnstructured(crd)
	if err != nil {
		return nil, fmt.Errorf("crd to unstructured: %w", err)
	}

	cleanVersions, found, err := unstructured.NestedSlice(clean.Object, "spec", "versions")
	if err != nil {
		return nil, fmt.Errorf("read versions: %w", err)
	}

	if !found {
		return clean, nil
	}

	rawVersions, _, err := unstructured.NestedSlice(raw, "spec", "versions")
	if err != nil {
		return nil, fmt.Errorf("read desired versions: %w", err)
	}

	for i, version := range cleanVersions {
		versionMap, ok := version.(map[string]any)
		if !ok {
			continue
		}

		name, _, _ := unstructured.NestedString(versionMap, "name")

		rawSchema, ok := findVersionSchema(rawVersions, name)
		if !ok {
			continue
		}

		props := &openapi.JSONSchemaProps{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(rawSchema, props); err != nil {
			return nil, fmt.Errorf("decode version %q schema: %w", name, err)
		}

		cleanSchema, err := runtime.DefaultUnstructuredConverter.ToUnstructured(props)
		if err != nil {
			return nil, fmt.Errorf("encode version %q schema: %w", name, err)
		}

		if err := unstructured.SetNestedMap(versionMap, cleanSchema, "schema", "openAPIV3Schema"); err != nil {
			return nil, fmt.Errorf("set version %q schema: %w", name, err)
		}

		cleanVersions[i] = versionMap
	}

	if err := unstructured.SetNestedSlice(clean.Object, cleanVersions, "spec", "versions"); err != nil {
		return nil, fmt.Errorf("set versions: %w", err)
	}

	return clean, nil
}

// findVersionSchema returns the openAPIV3Schema of the named version, matching by name
// rather than by index so it does not depend on the two documents agreeing on order.
func findVersionSchema(versions []any, name string) (map[string]any, bool) {
	for _, version := range versions {
		versionMap, ok := version.(map[string]any)
		if !ok {
			continue
		}

		if versionName, _, _ := unstructured.NestedString(versionMap, "name"); versionName != name {
			continue
		}

		schemaMap, found, err := unstructured.NestedMap(versionMap, "schema", "openAPIV3Schema")
		if err != nil || !found {
			return nil, false
		}

		return schemaMap, true
	}

	return nil, false
}

func (cp *CRDsInstaller) updateOrInsertCRD(ctx context.Context, crd *apiextensionsv1.CustomResourceDefinition, desired *unstructured.Unstructured) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		existing, err := cp.k8sClient.Resource(crdGVR).Get(ctx, crd.GetName(), apimachineryv1.GetOptions{})
		if apierrors.IsNotFound(err) {
			mergeLabels(desired, cp.crdExtraLabels)

			_, err = cp.k8sClient.Resource(crdGVR).Create(ctx, desired, apimachineryv1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("create crd: %w", err)
			}

			return nil
		}

		if err != nil {
			return fmt.Errorf("get crd from cluster: %w", err)
		}

		// typed view of the existing object is used only to reconcile storedVersions;
		// it is never written back as the CRD body (that would drop vendor extensions).
		existCRD := &apiextensionsv1.CustomResourceDefinition{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(existing.Object, existCRD); err != nil {
			return fmt.Errorf("existing crd from unstructured: %w", err)
		}

		versionsFromNewSpec := make(map[string]struct{}, len(crd.Spec.Versions))
		for _, version := range crd.Spec.Versions {
			versionsFromNewSpec[version.Name] = struct{}{}
		}

		newStoredVersions := make([]string, 0, len(versionsFromNewSpec))
		for _, version := range existCRD.Status.StoredVersions {
			if _, found := versionsFromNewSpec[version]; found {
				newStoredVersions = append(newStoredVersions, version)
			}
		}

		resourceVersion := existing.GetResourceVersion()
		if !slices.Equal(newStoredVersions, existCRD.Status.StoredVersions) {
			existCRD.Status.StoredVersions = newStoredVersions
			// the status subresource update ignores .spec, so writing the typed (lossy) body here is safe
			ucrd, err := utils.ToUnstructured(existCRD)
			if err != nil {
				return fmt.Errorf("crd to unstructured: %w", err)
			}

			resp, err := cp.k8sClient.Resource(crdGVR).Update(ctx, ucrd, apimachineryv1.UpdateOptions{}, "status")
			if err != nil {
				return fmt.Errorf("update existing crd status: %w", err)
			}

			if resp != nil {
				resourceVersion = resp.GetResourceVersion()
			}
		}

		// keep the in-cluster conversion webhook config (it is not present in the CRD file)
		if conv, found, err := unstructured.NestedFieldCopy(existing.Object, "spec", "conversion"); err == nil && found {
			if err := unstructured.SetNestedField(desired.Object, conv, "spec", "conversion"); err != nil {
				return fmt.Errorf("preserve conversion: %w", err)
			}
		}

		mergeLabels(desired, cp.crdExtraLabels)

		desiredSpec, _, err := unstructured.NestedMap(desired.Object, "spec")
		if err != nil {
			return fmt.Errorf("read desired spec: %w", err)
		}

		existingSpec, _, err := unstructured.NestedMap(existing.Object, "spec")
		if err != nil {
			return fmt.Errorf("read existing spec: %w", err)
		}

		// both sides are pruned to the same field set now — the desired spec by sanitize,
		// the existing one by the apiserver — so this can actually return early
		// ponytail: apiserver-defaulted .spec fields may differ from the manifest and cause
		// reconcile churn; the update stays idempotent, tighten the diff here if it ever churns.
		if cmp.Equal(existingSpec, desiredSpec) &&
			cmp.Equal(existing.GetLabels(), desired.GetLabels()) &&
			cmp.Equal(existing.GetAnnotations(), desired.GetAnnotations()) {
			return nil
		}

		// write back the fetched object with only spec/labels/annotations overlaid, so
		// server-managed metadata (finalizers, ownerReferences, uid, ...) is preserved.
		if err := unstructured.SetNestedMap(existing.Object, desiredSpec, "spec"); err != nil {
			return fmt.Errorf("set spec: %w", err)
		}
		existing.SetLabels(desired.GetLabels())
		existing.SetAnnotations(desired.GetAnnotations())
		existing.SetResourceVersion(resourceVersion)

		_, err = cp.k8sClient.Resource(crdGVR).Update(ctx, existing, apimachineryv1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("update crd: %w", err)
		}

		return nil
	})
}

func mergeLabels(u *unstructured.Unstructured, extra map[string]string) {
	labels := u.GetLabels()
	if labels == nil {
		labels = make(map[string]string, len(extra))
	}

	for k, v := range extra {
		labels[k] = v
	}

	u.SetLabels(labels)
}

func (cp *CRDsInstaller) GetCRDFromCluster(ctx context.Context, crdName string) (*apiextensionsv1.CustomResourceDefinition, error) {
	crd := &apiextensionsv1.CustomResourceDefinition{}

	o, err := cp.k8sClient.Resource(crdGVR).Get(ctx, crdName, apimachineryv1.GetOptions{})
	if err != nil {
		return nil, err
	}

	err = utils.FromUnstructured(o, &crd)
	if err != nil {
		return nil, err
	}

	return crd, nil
}

type InstallerOption func(*CRDsInstaller)

// NewCRDsInstaller creates new installer for CRDs
func NewCRDsInstaller(client dynamic.Interface, crdFilesPaths []string, options ...InstallerOption) *CRDsInstaller {
	i := &CRDsInstaller{
		k8sClient:     client,
		crdFilesPaths: crdFilesPaths,
		buffer:        make([]byte, bufSize),
		k8sTasks:      &multierror.Group{},
		appliedGVKs:   make([]string, 0),
	}

	for _, opt := range options {
		opt(i)
	}

	return i
}
