package objectpatch

import "github.com/deckhouse/module-sdk/pkg"

type CreateOption func(o pkg.PatchCollectorCreateOptionApplier)

func (opt CreateOption) Apply(o pkg.PatchCollectorCreateOptionApplier) {
	opt(o)
}

func CreateWithSubresource(subresource string) CreateOption {
	return func(o pkg.PatchCollectorCreateOptionApplier) {
		o.WithSubresource(subresource)
	}
}

func CreateWithIgnoreIfExists(ignore bool) CreateOption {
	return func(o pkg.PatchCollectorCreateOptionApplier) {
		o.WithIgnoreIfExists(ignore)
	}
}

func CreateWithUpdateIfExists(ignore bool) CreateOption {
	return func(o pkg.PatchCollectorCreateOptionApplier) {
		o.WithUpdateIfExists(ignore)
	}
}

type DeleteOption func(o pkg.PatchCollectorDeleteOptionApplier)

func (opt DeleteOption) Apply(o pkg.PatchCollectorDeleteOptionApplier) {
	opt(o)
}

func DeleteWithSubresource(subresource string) DeleteOption {
	return func(o pkg.PatchCollectorDeleteOptionApplier) {
		o.WithSubresource(subresource)
	}
}

type PatchOption func(o pkg.PatchCollectorPatchOptionApplier)

func (opt PatchOption) Apply(o pkg.PatchCollectorPatchOptionApplier) {
	opt(o)
}

func PatchWithSubresource(subresource string) PatchOption {
	return func(o pkg.PatchCollectorPatchOptionApplier) {
		o.WithSubresource(subresource)
	}
}

func PatchWithIgnoreMissingObject(ignore bool) PatchOption {
	return func(o pkg.PatchCollectorPatchOptionApplier) {
		o.WithIgnoreMissingObject(ignore)
	}
}

func PatchWithIgnoreHookError(ignore bool) PatchOption {
	return func(o pkg.PatchCollectorPatchOptionApplier) {
		o.WithIgnoreHookError(ignore)
	}
}

type FilterOption func(o pkg.PatchCollectorFilterOptionApplier)

func (opt FilterOption) Apply(o pkg.PatchCollectorFilterOptionApplier) {
	opt(o)
}

func FilterWithSubresource(subresource string) FilterOption {
	return func(o pkg.PatchCollectorFilterOptionApplier) {
		o.WithSubresource(subresource)
	}
}

func FilterWithIgnoreMissingObject(ignore bool) FilterOption {
	return func(o pkg.PatchCollectorFilterOptionApplier) {
		o.WithIgnoreMissingObject(ignore)
	}
}

func FilterWithIgnoreHookError(ignore bool) FilterOption {
	return func(o pkg.PatchCollectorFilterOptionApplier) {
		o.WithIgnoreHookError(ignore)
	}
}
