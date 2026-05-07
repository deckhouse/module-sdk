package framework

import (
	"sync"

	"github.com/deckhouse/module-sdk/pkg"
)

// PatchType identifies a recorded patch operation kind.
type PatchType string

const (
	PatchTypeCreate             PatchType = "Create"
	PatchTypeCreateOrUpdate     PatchType = "CreateOrUpdate"
	PatchTypeCreateIfNotExists  PatchType = "CreateIfNotExists"
	PatchTypeDelete             PatchType = "Delete"
	PatchTypeDeleteInBackground PatchType = "DeleteInBackground"
	PatchTypeDeleteNonCascading PatchType = "DeleteNonCascading"
	PatchTypeJSONPatch          PatchType = "JSONPatch"
	PatchTypeMergePatch         PatchType = "MergePatch"
	PatchTypeJQFilter           PatchType = "JQFilter"
)

// RecordedPatch is a structured copy of a single patch operation issued by
// a hook. It captures the operation type and all parameters so that tests
// can assert on the hook's intent.
type RecordedPatch struct {
	Type PatchType

	// For Create*: holds the runtime.Object / map / Unstructured.
	Object any

	// For Delete* / Patch* operations.
	APIVersion string
	Kind       string
	Namespace  string
	Name       string

	// For Patch operations.
	JSONPatch  any
	MergePatch any
	JQFilter   string

	// Original options as passed by the hook.
	Options []pkg.PatchCollectorOption
}

// recordingPatchCollector is a pkg.PatchCollector that records every call as
// a RecordedPatch. It also implements pkg.PatchCollectorOperation per record
// so that hooks see the operations they would normally see.
type recordingPatchCollector struct {
	mu      sync.Mutex
	records []RecordedPatch
}

func newRecordingPatchCollector() *recordingPatchCollector {
	return &recordingPatchCollector{records: make([]RecordedPatch, 0)}
}

var _ pkg.PatchCollector = (*recordingPatchCollector)(nil)

func (c *recordingPatchCollector) add(r RecordedPatch) {
	c.mu.Lock()
	c.records = append(c.records, r)
	c.mu.Unlock()
}

func (c *recordingPatchCollector) Records() []RecordedPatch {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]RecordedPatch, len(c.records))
	copy(out, c.records)
	return out
}

func (c *recordingPatchCollector) Operations() []pkg.PatchCollectorOperation {
	c.mu.Lock()
	defer c.mu.Unlock()
	ops := make([]pkg.PatchCollectorOperation, 0, len(c.records))
	for i := range c.records {
		ops = append(ops, &c.records[i])
	}
	return ops
}

// pkg.PatchCollectorOperation implementation.
func (r *RecordedPatch) Description() string { return string(r.Type) }
func (r *RecordedPatch) SetObjectPrefix(prefix string) {
	if prefix == "" || r.Name == "" {
		return
	}
	r.Name = prefix + "-" + r.Name
}

// === Create ===
func (c *recordingPatchCollector) Create(object any) {
	c.add(RecordedPatch{Type: PatchTypeCreate, Object: object})
}
func (c *recordingPatchCollector) CreateIfNotExists(object any) {
	c.add(RecordedPatch{Type: PatchTypeCreateIfNotExists, Object: object})
}
func (c *recordingPatchCollector) CreateOrUpdate(object any) {
	c.add(RecordedPatch{Type: PatchTypeCreateOrUpdate, Object: object})
}

// === Delete ===
func (c *recordingPatchCollector) Delete(apiVersion, kind, namespace, name string) {
	c.add(RecordedPatch{Type: PatchTypeDelete, APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name})
}
func (c *recordingPatchCollector) DeleteInBackground(apiVersion, kind, namespace, name string) {
	c.add(RecordedPatch{Type: PatchTypeDeleteInBackground, APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name})
}
func (c *recordingPatchCollector) DeleteNonCascading(apiVersion, kind, namespace, name string) {
	c.add(RecordedPatch{Type: PatchTypeDeleteNonCascading, APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name})
}

// === Patch ===
func (c *recordingPatchCollector) JSONPatch(jsonPatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
	c.add(RecordedPatch{Type: PatchTypeJSONPatch, APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name, JSONPatch: jsonPatch, Options: opts})
}
func (c *recordingPatchCollector) MergePatch(mergePatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
	c.add(RecordedPatch{Type: PatchTypeMergePatch, APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name, MergePatch: mergePatch, Options: opts})
}
func (c *recordingPatchCollector) JQFilter(jqfilter, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
	c.add(RecordedPatch{Type: PatchTypeJQFilter, APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name, JQFilter: jqfilter, Options: opts})
}
func (c *recordingPatchCollector) PatchWithJSON(jsonPatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
	c.add(RecordedPatch{Type: PatchTypeJSONPatch, APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name, JSONPatch: jsonPatch, Options: opts})
}
func (c *recordingPatchCollector) PatchWithMerge(mergePatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
	c.add(RecordedPatch{Type: PatchTypeMergePatch, APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name, MergePatch: mergePatch, Options: opts})
}
func (c *recordingPatchCollector) PatchWithJQ(jqfilter, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
	c.add(RecordedPatch{Type: PatchTypeJQFilter, APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name, JQFilter: jqfilter, Options: opts})
}
