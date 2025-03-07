package pkg

import (
	"github.com/tidwall/gjson"

	"github.com/deckhouse/module-sdk/pkg/utils"
)

type EMPatchCollector interface {
	PatchCollector
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

type PatchCollectorCreateOption interface {
	Apply(PatchCollectorCreateOptionApplier)
}

type PatchCollectorCreateOptionApplier interface {
	WithSubresource(subresource string)
}

type CreateOption func(o PatchCollectorCreateOptionApplier)

func (opt CreateOption) Apply(o PatchCollectorCreateOptionApplier) {
	opt(o)
}

func CreateWithSubresource(subresource string) CreateOption {
	return func(o PatchCollectorCreateOptionApplier) {
		o.WithSubresource(subresource)
	}
}

type PatchCollectorDeleteOption interface {
	Apply(PatchCollectorDeleteOptionApplier)
}

type PatchCollectorDeleteOptionApplier interface {
	WithSubresource(subresource string)
}

type DeleteOption func(o PatchCollectorDeleteOptionApplier)

func (opt DeleteOption) Apply(o PatchCollectorDeleteOptionApplier) {
	opt(o)
}

func DeleteWithSubresource(subresource string) DeleteOption {
	return func(o PatchCollectorDeleteOptionApplier) {
		o.WithSubresource(subresource)
	}
}

type PatchCollectorOption interface {
	Apply(PatchCollectorPatchOptionApplier)
}

type PatchCollectorPatchOptionApplier interface {
	WithSubresource(subresource string)
	WithIgnoreMissingObject(ignore bool)
	WithIgnoreHookError(update bool)
}

type PatchOption func(o PatchCollectorPatchOptionApplier)

func (opt PatchOption) Apply(o PatchCollectorPatchOptionApplier) {
	opt(o)
}

func PatchWithSubresource(subresource string) PatchOption {
	return func(o PatchCollectorPatchOptionApplier) {
		o.WithSubresource(subresource)
	}
}

func PatchWithIgnoreMissingObject(ignore bool) PatchOption {
	return func(o PatchCollectorPatchOptionApplier) {
		o.WithIgnoreMissingObject(ignore)
	}
}

func PatchWithIgnoreHookError(ignore bool) PatchOption {
	return func(o PatchCollectorPatchOptionApplier) {
		o.WithIgnoreHookError(ignore)
	}
}

type PatchCollectorFilterOption interface {
	Apply(PatchCollectorFilterOptionApplier)
}

type PatchCollectorFilterOptionApplier interface {
	WithSubresource(subresource string)
	WithIgnoreMissingObject(ignore bool)
	WithIgnoreHookError(update bool)
}

type FilterOption func(o PatchCollectorFilterOptionApplier)

func (opt FilterOption) Apply(o PatchCollectorFilterOptionApplier) {
	opt(o)
}

func FilterWithSubresource(subresource string) FilterOption {
	return func(o PatchCollectorFilterOptionApplier) {
		o.WithSubresource(subresource)
	}
}

func FilterWithIgnoreMissingObject(ignore bool) FilterOption {
	return func(o PatchCollectorFilterOptionApplier) {
		o.WithIgnoreMissingObject(ignore)
	}
}

func FilterWithIgnoreHookError(ignore bool) FilterOption {
	return func(o PatchCollectorFilterOptionApplier) {
		o.WithIgnoreHookError(ignore)
	}
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
