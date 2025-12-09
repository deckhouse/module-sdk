package pkg

import (
	"github.com/tidwall/gjson"

	"github.com/deckhouse/module-sdk/pkg/utils"
)

type OutputPatchCollector interface {
	// Deprecated: use PatchWithMerge instead
	PatchCollector
	Outputer
}

type OutputPatchableValuesCollector interface {
	PatchableValuesCollector
	Outputer
}

type PatchCollector interface {
	// object must be Unstructured, map[string]any or runtime.Object
	Create(object any)
	// object must be Unstructured, map[string]any or runtime.Object
	CreateIfNotExists(object any)
	// object must be Unstructured, map[string]any or runtime.Object
	CreateOrUpdate(object any)

	// The object exists in the key-value store until the garbage collector
	// deletes all the dependents whose ownerReference.blockOwnerDeletion=true
	// from the key-value store.  API sever will put the "foregroundDeletion"
	// finalizer on the object, and sets its deletionTimestamp.  This policy is
	// cascading, i.e., the dependents will be deleted with Foreground.
	Delete(apiVersion string, kind string, namespace string, name string)
	// Deletes the object from the key-value store, the garbage collector will
	// delete the dependents in the background.
	DeleteInBackground(apiVersion string, kind string, namespace string, name string)
	// Orphans the dependents.
	DeleteNonCascading(apiVersion string, kind string, namespace string, name string)

	// Deprecated: use PatchWithJSON instead
	JSONPatch(jsonPatch any, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorOption)
	// Deprecated: use PatchWithMerge instead
	MergePatch(mergePatch any, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorOption)
	// Deprecated: use PatchWithJQ instead
	JQFilter(jqfilter string, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorOption)

	// JSONPatch is a PatchType indicating the patch should be interpreted as a RFC6902 JSON Patch.
	// This patch format requires specifying operations, paths, and values explicitly.
	// See https://tools.ietf.org/html/rfc6902 for details.
	PatchWithJSON(jsonPatch any, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorOption)
	// MergePatch is a PatchType indicating the patch should be interpreted as a RFC7396 JSON Merge Patch.
	// This patch format replaces elements at the object level rather than requiring explicit operations.
	// See https://tools.ietf.org/html/rfc7396 for details.
	PatchWithMerge(mergePatch any, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorOption)
	// Mutate object with jq query
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

type ReadOnlyValuesCollector interface {
	ArrayCount(path string) (int, error)
	Exists(path string) bool
	Get(path string) gjson.Result
	GetOk(path string) (gjson.Result, bool)
	GetPatches() []*utils.ValuesPatchOperation
	GetRaw(path string) any
}
