package crdinstaller

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
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

	var errs error

	for {
		n, err := crdReader.Read(cp.buffer)
		if err != nil {
			if err == io.EOF {
				break
			}

			// the documents read so far may already have failed: reporting only the read
			// error would hide them
			return errors.Join(errs, err)
		}

		data := cp.buffer[:n]
		if len(data) == 0 {
			// some empty yaml document, or empty string before separator
			continue
		}

		rd := bytes.NewReader(data)

		// one bad document must not skip the rest of the file: the documents after it are
		// unrelated CRDs, and skipping them silently leaves the module without them
		if err := cp.putCRDToCluster(ctx, rd, n); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
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

	if err := applyServerDefaults(crd, desired); err != nil {
		return fmt.Errorf("default %s: %w", crd.Name, err)
	}

	// a schema this build cannot decode must not keep the CRD out of the cluster: the
	// document is queued as it came — what the installer did before it pruned anything —
	// and the error is reported afterwards. The apiserver prunes the key it does not
	// understand, while a missing CRD takes every custom resource of that kind with it.
	sanitizeErr := sanitize(desired)
	if sanitizeErr != nil {
		sanitizeErr = fmt.Errorf("sanitize %s: %w", crd.Name, sanitizeErr)
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

	return sanitizeErr
}

// sanitize drops the schema keys the apiserver does not know from the CRD document.
//
// x-doc-examples and friends, plus keys that look official but are not, like
// x-kubernetes-immutable, are removed here instead of being sent. The apiserver prunes
// them anyway, after logging one "unknown field" warning per occurrence, and because it
// prunes them the stored spec could never equal the desired one, so every run issued a
// pointless Update.
//
// Only the version schemas are rewritten; the rest of the document is passed through
// untouched. Rebuilding it from apiextensionsv1.CustomResourceDefinition instead would
// pin the installer to the CRD fields of the apiextensions-apiserver it was compiled
// against, and silently strip anything a newer or patched apiserver understands.
func sanitize(desired *unstructured.Unstructured) error {
	versions, ok := nestedValue(desired.Object, "spec", "versions").([]any)
	if !ok {
		// no versions, or not a list: let the apiserver reject the document
		return nil
	}

	for _, version := range versions {
		versionMap, ok := version.(map[string]any)
		if !ok {
			continue
		}

		schema, ok := versionMap["schema"].(map[string]any)
		if !ok {
			continue
		}

		rawSchema, ok := schema["openAPIV3Schema"].(map[string]any)
		if !ok {
			continue
		}

		cleanSchema, err := openapi.Prune(rawSchema)
		if err != nil {
			name, _, _ := unstructured.NestedString(versionMap, "name")

			return fmt.Errorf("version %q schema: %w", name, err)
		}

		schema["openAPIV3Schema"] = cleanSchema
	}

	return nil
}

// nestedValue returns the value at the given path without copying it, or nil if the path
// does not lead to one. Errors are the same as absence for every caller here.
func nestedValue(obj map[string]any, fields ...string) any {
	value, found, err := unstructured.NestedFieldNoCopy(obj, fields...)
	if err != nil || !found {
		return nil
	}

	return value
}

// applyServerDefaults writes the .spec fields the apiserver fills in itself into the
// document, so a manifest that omits them does not differ from the stored object on
// every single reconcile.
//
// ponytail: only the .spec fields known to churn are written here; if a CRD still updates
// on every reconcile, diff the stored object against the manifest and add the field that
// differs.
func applyServerDefaults(crd *apiextensionsv1.CustomResourceDefinition, desired *unstructured.Unstructured) error {
	apiextensionsv1.SetDefaults_CustomResourceDefinitionSpec(&crd.Spec)

	// served and storage have no omitempty upstream, so the stored object always carries
	// both on every version — a manifest that omits either would differ from it forever
	versions, _ := nestedValue(desired.Object, "spec", "versions").([]any)
	for _, version := range versions {
		versionMap, ok := version.(map[string]any)
		if !ok {
			continue
		}

		for _, field := range []string{"served", "storage"} {
			if _, ok := versionMap[field]; !ok {
				versionMap[field] = false
			}
		}
	}

	// spec.conversion is defaulted too, but updateOrInsertCRD always takes the in-cluster
	// one, which the apiserver has already defaulted
	names := map[string]string{
		"singular": crd.Spec.Names.Singular,
		"listKind": crd.Spec.Names.ListKind,
	}

	for field, value := range names {
		if value == "" {
			continue
		}

		if err := unstructured.SetNestedField(desired.Object, value, "spec", "names", field); err != nil {
			return fmt.Errorf("set spec.names.%s: %w", field, err)
		}
	}

	return nil
}

func (cp *CRDsInstaller) updateOrInsertCRD(ctx context.Context, crd *apiextensionsv1.CustomResourceDefinition, desired *unstructured.Unstructured) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		desired.SetLabels(overlay(desired.GetLabels(), cp.crdExtraLabels))

		existing, err := cp.k8sClient.Resource(crdGVR).Get(ctx, crd.GetName(), apimachineryv1.GetOptions{})
		if apierrors.IsNotFound(err) {
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

		desiredSpec, _, err := unstructured.NestedMap(desired.Object, "spec")
		if err != nil {
			return fmt.Errorf("read desired spec: %w", err)
		}

		existingSpec, _, err := unstructured.NestedMap(existing.Object, "spec")
		if err != nil {
			return fmt.Errorf("read existing spec: %w", err)
		}

		// labels and annotations belong to whoever wrote them: dropping
		// app.kubernetes.io/managed-by or meta.helm.sh/release-name breaks the next helm
		// upgrade of the chart that installed the CRD, so both maps are overlaid, never
		// replaced. The result is what the object must end up holding, which is also what
		// makes the comparison below exact.
		//
		// ponytail: overlaying means a key the manifest stops declaring stays in the
		// cluster — retracting one needs the field ownership the apiserver keeps for
		// server-side apply; switch this whole update to Apply if that ever matters.
		labels := overlay(existing.GetLabels(), desired.GetLabels())
		annotations := overlay(existing.GetAnnotations(), desired.GetAnnotations())

		// The specs are compared, not merged: the desired one is pruned by sanitize and the
		// existing one by the apiserver, so a manifest applied over the state the apiserver
		// derived from it is a no-op. A .spec key this cluster's apiserver does not know is
		// the exception — it prunes the key while the desired document keeps it, so that CRD
		// is updated on every reconcile. Deliberate: the key is sent so an apiserver that
		// does know it gets it.
		if cmp.Equal(existingSpec, desiredSpec) &&
			cmp.Equal(existing.GetLabels(), labels) &&
			cmp.Equal(existing.GetAnnotations(), annotations) {
			return nil
		}

		// write back the fetched object with only spec/labels/annotations overlaid, so
		// server-managed metadata (finalizers, ownerReferences, uid, ...) is preserved.
		if err := unstructured.SetNestedMap(existing.Object, desiredSpec, "spec"); err != nil {
			return fmt.Errorf("set spec: %w", err)
		}
		existing.SetLabels(labels)
		existing.SetAnnotations(annotations)
		existing.SetResourceVersion(resourceVersion)

		_, err = cp.k8sClient.Resource(crdGVR).Update(ctx, existing, apimachineryv1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("update crd: %w", err)
		}

		return nil
	})
}

// overlay returns base with over's keys written on top of it. A key base holds and over
// does not is kept.
//
// nil in, nil out: the apiserver returns no map at all for empty labels or annotations, and
// an empty one written back would never compare equal to that — SetLabels/SetAnnotations
// remove the field for nil and set metadata.labels: {} for an empty map.
func overlay(base, over map[string]string) map[string]string {
	if len(over) == 0 {
		return base
	}

	out := make(map[string]string, len(base)+len(over))
	maps.Copy(out, base)
	maps.Copy(out, over)

	return out
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
