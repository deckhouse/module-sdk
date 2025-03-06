package objectpatch

import "github.com/deckhouse/module-sdk/pkg"

type Patch struct {
	patchValues map[string]any
}

func (p *Patch) Operation() string {
	return p.patchValues["operation"].(string)
}

func (p *Patch) ApplyCreateOptions(lo *pkg.PatchCollectorCreateOptions) {
	p.patchValues["subresource"] = lo.Subresource
}

func (p *Patch) ApplyDeleteOptions(lo *pkg.PatchCollectorDeleteOptions) {
	p.patchValues["subresource"] = lo.Subresource
}

func (p *Patch) ApplyPatchOptions(lo *pkg.PatchCollectorPatchOptions) {
	p.patchValues["subresource"] = lo.Subresource
	p.patchValues["ignoreMissingObjects"] = lo.IgnoreMissingObjects
}
