package objectpatch

import "github.com/deckhouse/module-sdk/pkg"

type PatchOption func(o pkg.PatchCollectorOptionApplier)

func (opt PatchOption) Apply(o pkg.PatchCollectorOptionApplier) {
	opt(o)
}

func WithSubresource(subresource string) PatchOption {
	return func(o pkg.PatchCollectorOptionApplier) {
		o.WithSubresource(subresource)
	}
}

func WithIgnoreMissingObject(ignore bool) PatchOption {
	return func(o pkg.PatchCollectorOptionApplier) {
		o.WithIgnoreMissingObject(ignore)
	}
}

func WithIgnoreHookError(ignore bool) PatchOption {
	return func(o pkg.PatchCollectorOptionApplier) {
		o.WithIgnoreHookError(ignore)
	}
}
