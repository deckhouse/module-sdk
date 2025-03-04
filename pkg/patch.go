package pkg

import (
	"github.com/tidwall/gjson"

	"github.com/deckhouse/module-sdk/pkg/utils"
)

type PatchCollector interface {
	Outputer

	Create(data any, opts ...PatchCollectorCreateOption)
	CreateIfNotExists(data any, opts ...PatchCollectorCreateOption)
	CreateOrUpdate(data any, opts ...PatchCollectorCreateOption)

	Delete(apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorDeleteOption)
	DeleteInBackground(apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorDeleteOption)
	DeleteNonCascading(apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorDeleteOption)

	JQPatch(filter string, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorPatchOption)
	MergePatch(patch any, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorPatchOption)
	JSONPatch(patch []any, apiVersion string, kind string, namespace string, name string, opts ...PatchCollectorPatchOption)
}

type PatchCollectorCreateOption interface {
	Apply(optsApplier PatchCollectorCreateOptionApplier)
}

type PatchCollectorCreateOptionApplier interface {
	WithSubresource(subresource string)
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
