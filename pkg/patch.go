package pkg

import (
	"github.com/tidwall/gjson"

	"github.com/deckhouse/module-sdk/pkg/utils"
)

type EMPatchCollector interface {
	PatchCollector
	Outputer
}

type EMPatchableValuesCollector interface {
	PatchableValuesCollector
	Outputer
}

type PatchCollector interface {
	Create(object any)
	CreateIfNotExists(object any)
	CreateOrUpdate(object any)

	Delete(apiVersion string, kind string, namespace string, name string)
	DeleteInBackground(apiVersion string, kind string, namespace string, name string)
	DeleteNonCascading(apiVersion string, kind string, namespace string, name string)

	// deprecated use PatchWithJSON instead
	JSONPatch(jsonPatch any, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorOption)
	// deprecated use PatchWithMerge instead
	MergePatch(mergePatch any, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorOption)
	// deprecated use PatchWithJQ instead
	JQFilter(jqfilter string, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorOption)

	PatchWithJSON(jsonPatch any, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorOption)
	PatchWithMerge(mergePatch any, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorOption)
	PatchWithJQ(jqfilter string, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorOption)

	Operations() []PatchCollectorOperation
}

// There are 4 types of operations:
//
// - createOperation to create or update object via Create and Update API calls. Unstructured, map[string]any or runtime.Object is required.
//
// - deleteOperation to delete object via Delete API call
//
// - patchOperation to modify object via Patch API call
//
// - filterOperation to modify object via Get-filter-Update process
type PatchCollectorOperation interface {
	Description() string
}

type PatchCollectorOption interface {
	Apply(PatchCollectorOptionApplier)
}

type PatchCollectorOptionApplier interface {
	WithSubresource(subresource string)
	WithIgnoreMissingObject(ignore bool)
	WithIgnoreHookError(update bool)
}

type PatchableValuesCollector interface {
	ArrayCount(path string) (int, error)
	Exists(path string) bool
	Get(path string) gjson.Result
	GetOk(path string) (gjson.Result, bool)
	GetPatches() []*utils.ValuesPatchOperation
	GetRaw(path string) any
	Remove(path string)
	Set(path string, value any)
}
