package objectpatch

import "github.com/deckhouse/module-sdk/pkg"

var _ pkg.PatchCollectorCreateOptionApplier = (*Patch)(nil)
var _ pkg.PatchCollectorDeleteOptionApplier = (*Patch)(nil)
var _ pkg.PatchCollectorPatchOptionApplier = (*Patch)(nil)
var _ pkg.PatchCollectorFilterOptionApplier = (*Patch)(nil)

type Patch struct {
	patchValues map[string]any
}

func (p *Patch) Description() string {
	return p.patchValues["operation"].(string)
}

func (p *Patch) WithSubresource(subresource string) {
	p.patchValues["subresource"] = subresource
}

func (p *Patch) WithIgnoreMissingObject(ignore bool) {
	p.patchValues["ignoreMissingObjects"] = ignore
}

func (p *Patch) WithIgnoreHookError(ignore bool) {
	p.patchValues["ignoreHookError"] = ignore
}

func (p *Patch) WithIgnoreIfExists(ignore bool) {
	p.patchValues["ignoreIfExists"] = ignore
}

func (p *Patch) WithUpdateIfExists(update bool) {
	p.patchValues["updateIfExists"] = update
}
