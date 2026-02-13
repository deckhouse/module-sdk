package objectpatch

import "github.com/deckhouse/module-sdk/pkg"

// PatchOption is a functional option for configuring patch operations.
type PatchOption func(o pkg.PatchCollectorOptionApplier)

// Apply implements pkg.PatchCollectorOption interface.
func (opt PatchOption) Apply(o pkg.PatchCollectorOptionApplier) {
	opt(o)
}

// WithSubresource targets a specific subresource (e.g., "status", "scale").
func WithSubresource(subresource string) PatchOption {
	return func(o pkg.PatchCollectorOptionApplier) {
		o.WithSubresource(subresource)
	}
}

// WithIgnoreMissingObject prevents errors when the target object doesn't exist.
func WithIgnoreMissingObject(ignore bool) PatchOption {
	return func(o pkg.PatchCollectorOptionApplier) {
		o.WithIgnoreMissingObject(ignore)
	}
}

// WithIgnoreHookError allows hook execution to continue even if this patch fails.
func WithIgnoreHookError(ignore bool) PatchOption {
	return func(o pkg.PatchCollectorOptionApplier) {
		o.WithIgnoreHookError(ignore)
	}
}
