package crdinstaller

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apimachineryYaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"

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
	var crd *apiextensionsv1.CustomResourceDefinition

	err := apimachineryYaml.NewYAMLOrJSONDecoder(crdReader, bufferSize).Decode(&crd)
	if err != nil {
		return err
	}

	// it could be a comment or some other peace of yaml file, skip it
	if crd == nil {
		return nil
	}

	if crd.APIVersion != apiextensionsv1.SchemeGroupVersion.String() && crd.Kind != "CustomResourceDefinition" {
		return fmt.Errorf("invalid CRD document apiversion/kind: '%s/%s'", crd.APIVersion, crd.Kind)
	}

	if len(crd.ObjectMeta.Labels) == 0 {
		crd.ObjectMeta.Labels = make(map[string]string, 1)
	}

	for crdExtraLabel := range cp.crdExtraLabels {
		crd.ObjectMeta.Labels[crdExtraLabel] = cp.crdExtraLabels[crdExtraLabel]
	}

	cp.k8sTasks.Go(func() error {
		err := cp.updateOrInsertCRD(ctx, crd)
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

func (cp *CRDsInstaller) updateOrInsertCRD(ctx context.Context, crd *apiextensionsv1.CustomResourceDefinition) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		existCRD, err := cp.GetCRDFromCluster(ctx, crd.GetName())
		if apierrors.IsNotFound(err) {
			ucrd, err := utils.ToUnstructured(crd)
			if err != nil {
				return err
			}

			_, err = cp.k8sClient.Resource(crdGVR).Create(ctx, ucrd, apimachineryv1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("create crd: %w", err)
			}

			return nil
		}

		if err != nil {
			return fmt.Errorf("get crd from cluster: %w", err)
		}

		if existCRD.Spec.Conversion != nil {
			crd.Spec.Conversion = existCRD.Spec.Conversion
		}

		if cmp.Equal(existCRD.Spec, crd.Spec) &&
			cmp.Equal(existCRD.GetLabels(), crd.GetLabels()) &&
			cmp.Equal(existCRD.GetAnnotations(), crd.GetAnnotations()) {
			return nil
		}

		existCRD.Spec = crd.Spec
		existCRD.Labels = crd.Labels
		existCRD.Annotations = crd.Annotations

		if len(existCRD.ObjectMeta.Labels) == 0 {
			existCRD.ObjectMeta.Labels = make(map[string]string, 1)
		}

		for crdExtraLabel := range cp.crdExtraLabels {
			existCRD.ObjectMeta.Labels[crdExtraLabel] = cp.crdExtraLabels[crdExtraLabel]
		}

		ucrd, err := utils.ToUnstructured(existCRD)
		if err != nil {
			return err
		}

		_, err = cp.k8sClient.Resource(crdGVR).Update(ctx, ucrd, apimachineryv1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("update crd: %w", err)
		}

		return nil
	})
}

func (cp *CRDsInstaller) GetCRDFromCluster(ctx context.Context, crdName string) (*apiextensionsv1.CustomResourceDefinition, error) {
	crd := &v1.CustomResourceDefinition{}

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
