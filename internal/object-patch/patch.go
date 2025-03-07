package objectpatch

import "github.com/deckhouse/module-sdk/pkg"

var _ pkg.PatchCollectorOptionApplier = (*Patch)(nil)

type operationKind int

const (
	operationCreate operationKind = iota
	operationDelete
	operationPatch
	operationFilter
)

type Patch struct {
	kind        operationKind
	patchValues map[string]any
}

func (p *Patch) Description() string {
	return p.patchValues["operation"].(string)
}

func (p *Patch) Kind() int {
	return int(p.kind)
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
