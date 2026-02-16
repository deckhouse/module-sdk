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

// GetName returns the name of the object to patch.
func (p *Patch) GetName() string {
	name, ok := p.patchValues["name"]
	if !ok {
		return ""
	}

	// Handle both string and typed names
	return fmt.Sprintf("%v", name)
}

// GetNamespace returns the namespace of the object to patch.
func (p *Patch) GetNamespace() string {
	ns, ok := p.patchValues["namespace"]
	if !ok {
		return ""
	}

	// Handle both string and typed namespaces
	return fmt.Sprintf("%v", ns)
}

// SetPrifixName sets the name for the patch operation with a prefix.
func (p *Patch) SetNamePrefix(prefix string) {
	// Set the name for the patch operation with a prefix.
	// This is used to identify the target object in Kubernetes.
	if p.patchValues == nil {
		p.patchValues = make(map[string]any)
	}
	p.patchValues["name"] = fmt.Sprintf("%s-%s", prefix, p.GetName())
}

// SetName sets the name for the patch operation.
func (p *Patch) SetName(name string) {
	// Set the name for the patch operation.
	// This is used to identify the target object in Kubernetes.
	if p.patchValues == nil {
		p.patchValues = make(map[string]any)
	}
	p.patchValues["name"] = name
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
