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
	Create(object any, opts ...PatchCollectorCreateOption)
	CreateIfNotExists(object any, opts ...PatchCollectorCreateOption)
	CreateOrUpdate(object any, opts ...PatchCollectorCreateOption)

	Delete(apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorDeleteOption)
	DeleteInBackground(apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorDeleteOption)
	DeleteNonCascading(apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorDeleteOption)

	JQFilter(jqfilter string, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorFilterOption)

	JSONPatch(jsonPatch any, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorPatchOption)
	MergePatch(mergePatch any, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorPatchOption)

	Operations() []PatchCollectorOperation
}

type PatchCollectorOperation interface {
	Description() string
}

type PatchCollectorCreateOption interface {
	Apply(PatchCollectorCreateOptionApplier)
}

type PatchCollectorCreateOptionApplier interface {
	WithSubresource(subresource string)
	WithIgnoreIfExists(ignore bool)
	WithUpdateIfExists(update bool)
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

func CreateWithIgnoreIfExists(ignore bool) CreateOption {
	return func(o PatchCollectorCreateOptionApplier) {
		o.WithIgnoreIfExists(ignore)
	}
}

func CreateWithUpdateIfExists(ignore bool) CreateOption {
	return func(o PatchCollectorCreateOptionApplier) {
		o.WithUpdateIfExists(ignore)
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

type PatchCollectorPatchOption interface {
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
