package helpers

import (
	"sync"

	"github.com/deckhouse/module-sdk/pkg"
)

// RecordedOp captures the parameters of a single PatchCollector call.
//
// The fields are populated only for the operation that is relevant for
// the type:
//
//   - Op = "Create" / "CreateOrUpdate" / "CreateIfNotExists" → Object
//   - Op = "Delete*"                                         → APIVersion, Kind, Namespace, Name
//   - Op = "JSONPatch" / "MergePatch" / "JQFilter"           → APIVersion, Kind, Namespace, Name, Patch
//
// RecordedOp also implements pkg.PatchCollectorOperation so it can be used
// in code paths that expect that interface.
type RecordedOp struct {
	Op string

	Object any

	APIVersion string
	Kind       string
	Namespace  string
	Name       string

	Patch    any
	JQFilter string

	Options []pkg.PatchCollectorOption
}

// Description implements pkg.PatchCollectorOperation.
func (r *RecordedOp) Description() string { return r.Op }

// SetObjectPrefix implements pkg.PatchCollectorOperation. It mirrors the
// real implementation by prefixing the recorded Name when set.
func (r *RecordedOp) SetObjectPrefix(prefix string) {
	if prefix == "" || r.Name == "" {
		return
	}
	r.Name = prefix + "-" + r.Name
}

// RecordingPatchCollector is a pkg.PatchCollector that records every call
// and exposes the recorded operations for assertions.
//
// It is intentionally simple — no replay against a fake cluster, no
// validation. For full end-to-end testing prefer testing/framework.
type RecordingPatchCollector struct {
	mu  sync.Mutex
	ops []*RecordedOp
}

// NewRecordingPatchCollector returns an empty RecordingPatchCollector.
func NewRecordingPatchCollector() *RecordingPatchCollector {
	return &RecordingPatchCollector{}
}

var _ pkg.PatchCollector = (*RecordingPatchCollector)(nil)

// Recorded returns a copy of the recorded operations in the order they
// were issued by the hook.
func (c *RecordingPatchCollector) Recorded() []*RecordedOp {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]*RecordedOp, len(c.ops))
	copy(out, c.ops)
	return out
}

// Operations implements pkg.PatchCollector.
func (c *RecordingPatchCollector) Operations() []pkg.PatchCollectorOperation {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]pkg.PatchCollectorOperation, 0, len(c.ops))
	for _, op := range c.ops {
		out = append(out, op)
	}
	return out
}

// Filter returns the subset of recorded operations whose Op equals one of
// the provided values. It is a small convenience for assertions:
//
//	deletes := pc.Filter("Delete", "DeleteInBackground")
func (c *RecordingPatchCollector) Filter(ops ...string) []*RecordedOp {
	allowed := make(map[string]struct{}, len(ops))
	for _, o := range ops {
		allowed[o] = struct{}{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]*RecordedOp, 0)
	for _, op := range c.ops {
		if _, ok := allowed[op.Op]; ok {
			out = append(out, op)
		}
	}
	return out
}

func (c *RecordingPatchCollector) record(op *RecordedOp) {
	c.mu.Lock()
	c.ops = append(c.ops, op)
	c.mu.Unlock()
}

// Create implements pkg.PatchCollector.
func (c *RecordingPatchCollector) Create(object any) {
	c.record(&RecordedOp{Op: "Create", Object: object})
}

// CreateIfNotExists implements pkg.PatchCollector.
func (c *RecordingPatchCollector) CreateIfNotExists(object any) {
	c.record(&RecordedOp{Op: "CreateIfNotExists", Object: object})
}

// CreateOrUpdate implements pkg.PatchCollector.
func (c *RecordingPatchCollector) CreateOrUpdate(object any) {
	c.record(&RecordedOp{Op: "CreateOrUpdate", Object: object})
}

// Delete implements pkg.PatchCollector.
func (c *RecordingPatchCollector) Delete(apiVersion, kind, namespace, name string) {
	c.record(&RecordedOp{Op: "Delete", APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name})
}

// DeleteInBackground implements pkg.PatchCollector.
func (c *RecordingPatchCollector) DeleteInBackground(apiVersion, kind, namespace, name string) {
	c.record(&RecordedOp{Op: "DeleteInBackground", APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name})
}

// DeleteNonCascading implements pkg.PatchCollector.
func (c *RecordingPatchCollector) DeleteNonCascading(apiVersion, kind, namespace, name string) {
	c.record(&RecordedOp{Op: "DeleteNonCascading", APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name})
}

// JSONPatch implements pkg.PatchCollector (deprecated alias).
func (c *RecordingPatchCollector) JSONPatch(jsonPatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
	c.record(&RecordedOp{Op: "JSONPatch", APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name, Patch: jsonPatch, Options: opts})
}

// MergePatch implements pkg.PatchCollector (deprecated alias).
func (c *RecordingPatchCollector) MergePatch(mergePatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
	c.record(&RecordedOp{Op: "MergePatch", APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name, Patch: mergePatch, Options: opts})
}

// JQFilter implements pkg.PatchCollector (deprecated alias).
func (c *RecordingPatchCollector) JQFilter(jqfilter, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
	c.record(&RecordedOp{Op: "JQFilter", APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name, JQFilter: jqfilter, Options: opts})
}

// PatchWithJSON implements pkg.PatchCollector.
func (c *RecordingPatchCollector) PatchWithJSON(jsonPatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
	c.record(&RecordedOp{Op: "JSONPatch", APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name, Patch: jsonPatch, Options: opts})
}

// PatchWithMerge implements pkg.PatchCollector.
func (c *RecordingPatchCollector) PatchWithMerge(mergePatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
	c.record(&RecordedOp{Op: "MergePatch", APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name, Patch: mergePatch, Options: opts})
}

// PatchWithJQ implements pkg.PatchCollector.
func (c *RecordingPatchCollector) PatchWithJQ(jqfilter, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
	c.record(&RecordedOp{Op: "JQFilter", APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name, JQFilter: jqfilter, Options: opts})
}
