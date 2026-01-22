package objectpatch

import (
	"fmt"

	"github.com/deckhouse/module-sdk/pkg"
)

// Compile-time interface compliance check
var _ pkg.PatchCollectorOptionApplier = (*Patch)(nil)

// Patch represents a single patch operation with its parameters stored as a map.
// The patchValues map is serialized to JSON when sent to the shell-operator.
type Patch struct {
	patchValues map[string]any
}

// Description returns a human-readable description of the patch operation.
// Returns "unknown" if operation type is missing or invalid.
func (p *Patch) Description() string {
	op, ok := p.patchValues["operation"]
	if !ok {
		return "unknown"
	}

	// Handle both string and typed operation enums (CreateOperation, etc.)
	return fmt.Sprintf("%v", op)
}

// WithSubresource sets the subresource to patch (e.g., "status", "scale").
func (p *Patch) WithSubresource(subresource string) {
	p.patchValues["subresource"] = subresource
}

// WithIgnoreMissingObject prevents errors when the target object doesn't exist.
func (p *Patch) WithIgnoreMissingObject(ignore bool) {
	p.patchValues["ignoreMissingObjects"] = ignore
}

// WithIgnoreHookError continues execution even if this patch fails.
func (p *Patch) WithIgnoreHookError(ignore bool) {
	p.patchValues["ignoreHookError"] = ignore
}
