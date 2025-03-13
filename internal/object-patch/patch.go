package objectpatch

import "github.com/deckhouse/module-sdk/pkg"

var _ pkg.PatchCollectorOptionApplier = (*Patch)(nil)

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
