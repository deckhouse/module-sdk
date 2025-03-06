package objectpatch

import (
	"github.com/deckhouse/module-sdk/pkg"
)

func WithSubresource(str string) Subresource {
	return Subresource(str)
}

// Subresource is a configuration option for specifying a subresource.
type Subresource string

// ApplyToCreate applies this configuration to the given list options.
func (m Subresource) ApplyToCreate(opts *pkg.PatchCollectorCreateOptions) {
	opts.Subresource = string(m)
}

// ApplyToDelete applies this configuration to the given list options.
func (m Subresource) ApplyToDelete(opts *pkg.PatchCollectorDeleteOptions) {
	opts.Subresource = string(m)
}

// ApplyToPatch applies this configuration to the given list options.
func (m Subresource) ApplyToPatch(opts *pkg.PatchCollectorPatchOptions) {
	opts.Subresource = string(m)
}

func WithIgnoreMissingObject(ignore bool) IgnoreMissingObjects {
	return IgnoreMissingObjects(ignore)
}

// Subresource is a configuration option for specifying a subresource.
type IgnoreMissingObjects bool

// ApplyToPatch applies this configuration to the given list options.
func (m IgnoreMissingObjects) ApplyToPatch(opts *pkg.PatchCollectorPatchOptions) {
	opts.IgnoreMissingObjects = bool(m)
}
