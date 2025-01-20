package objectpatch

import "github.com/deckhouse/module-sdk/pkg"

var _ pkg.PatchCollectorPatchOptionApplier = (Patch)(nil)
var _ pkg.PatchCollectorDeleteOptionApplier = (Patch)(nil)

type Patch map[string]any

func (p Patch) Operation() string {
	return p["operation"].(string)
}

func (p Patch) WithSubresource(subresource string) {
	p["subresource"] = subresource
}

func (p Patch) WithIgnoreMissingObjects(ignore bool) {
	p["ignoreMissingObjects"] = ignore
}
