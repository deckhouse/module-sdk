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
	ApplyToCreate(*PatchCollectorCreateOptions)
}

type PatchCollectorCreateOptions struct {
	Subresource string
}

// ApplyOptions applies the given list options on these options,
// and then returns itself (for convenient chaining).
func (o *PatchCollectorCreateOptions) ApplyOptions(opts []PatchCollectorCreateOption) *PatchCollectorCreateOptions {
	for _, opt := range opts {
		opt.ApplyToCreate(o)
	}

	return o
}

type PatchCollectorDeleteOption interface {
	ApplyToDelete(*PatchCollectorDeleteOptions)
}

type PatchCollectorDeleteOptions struct {
	Subresource string
}

// ApplyOptions applies the given list options on these options,
// and then returns itself (for convenient chaining).
func (o *PatchCollectorDeleteOptions) ApplyOptions(opts []PatchCollectorDeleteOption) *PatchCollectorDeleteOptions {
	for _, opt := range opts {
		opt.ApplyToDelete(o)
	}

	return o
}

type PatchCollectorPatchOption interface {
	ApplyToPatch(*PatchCollectorPatchOptions)
}

type PatchCollectorPatchOptions struct {
	Subresource          string
	IgnoreMissingObjects bool
}

// ApplyOptions applies the given list options on these options,
// and then returns itself (for convenient chaining).
func (o *PatchCollectorPatchOptions) ApplyOptions(opts []PatchCollectorPatchOption) *PatchCollectorPatchOptions {
	for _, opt := range opts {
		opt.ApplyToPatch(o)
	}

	return o
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
