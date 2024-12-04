package pkg

import (
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/deckhouse/module-sdk/pkg/utils"
)

type PatchCollector interface {
	Outputer

	Create(data *unstructured.Unstructured)
	CreateIfNotExists(data *unstructured.Unstructured)
	CreateOrUpdate(data *unstructured.Unstructured)

	Delete(apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorDeleteOption)
	DeleteInBackground(apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorDeleteOption)
	DeleteNonCascading(apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorDeleteOption)

	JQPatch(filter string, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorPatchOption)
	MergePatch(patch any, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorPatchOption)
	JSONPatch(patch []any, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorPatchOption)
}

type PatchCollectorDeleteOption interface {
	Apply(optsApplier PatchCollectorDeleteOptionApplier)
}

type PatchCollectorDeleteOptionApplier interface {
	WithSubresource(subresource string)
}

type PatchCollectorPatchOption interface {
	Apply(optsApplier PatchCollectorPatchOptionApplier)
}

type PatchCollectorPatchOptionApplier interface {
	WithSubresource(subresource string)
	WithIgnoreMissingObjects(ignore bool)
}

type PatchableValuesCollector interface {
	Outputer

	ArrayCount(path string) (int, error)
	Exists(path string) bool
	Get(path string) gjson.Result
	GetOk(path string) (gjson.Result, bool)
	GetPatches() []*utils.ValuesPatchOperation
	GetRaw(path string) any
	Remove(path string)
	Set(path string, value any)
}
