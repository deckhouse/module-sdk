package objectpatch

import "github.com/deckhouse/module-sdk/pkg"

type CreatePatchOption func(optionsApplier pkg.PatchCollectorCreateOptionApplier)

func (opt CreatePatchOption) Apply(o pkg.PatchCollectorCreateOptionApplier) {
	opt(o)
}

func CreateWithSubresource(subresource string) CreatePatchOption {
	return func(optionsApplier pkg.PatchCollectorCreateOptionApplier) {
		optionsApplier.WithSubresource(subresource)
	}
}

type DeletePatchOption func(optionsApplier pkg.PatchCollectorDeleteOptionApplier)

func (opt DeletePatchOption) Apply(o pkg.PatchCollectorDeleteOptionApplier) {
	opt(o)
}

func DeleteWithSubresource(subresource string) DeletePatchOption {
	return func(optionsApplier pkg.PatchCollectorDeleteOptionApplier) {
		optionsApplier.WithSubresource(subresource)
	}
}

type PatchPatchOption func(optionsApplier pkg.PatchCollectorPatchOptionApplier)

func (opt PatchPatchOption) Apply(o pkg.PatchCollectorPatchOptionApplier) {
	opt(o)
}

func PatchWithSubresource(subresource string) PatchPatchOption {
	return func(optionsApplier pkg.PatchCollectorPatchOptionApplier) {
		optionsApplier.WithSubresource(subresource)
	}
}

func PatchWithIgnoreMissingObjects(ignore bool) PatchPatchOption {
	return func(optionsApplier pkg.PatchCollectorPatchOptionApplier) {
		optionsApplier.WithIgnoreMissingObjects(ignore)
	}
}
