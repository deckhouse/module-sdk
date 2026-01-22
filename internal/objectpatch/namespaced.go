package objectpatch

import (
	"io"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/utils"
)

// Compile-time interface compliance check
var _ pkg.NamespacedPatchCollector = (*NamespacedPatchCollector)(nil)

// NamespacedPatchCollector wraps PatchCollector to automatically inject
// namespace into all operations. Used for application hooks where the namespace
// is fixed and should not be specified by the caller.
// Note: This collector is not thread-safe; do not use concurrently.
type NamespacedPatchCollector struct {
	namespace string
	collector *PatchCollector
}

func NewNamespacedCollector(namespace string, logger *log.Logger) *NamespacedPatchCollector {
	return &NamespacedPatchCollector{
		namespace: namespace,
		collector: NewCollector(logger),
	}
}

// Create creates the object in the cluster.
func (c *NamespacedPatchCollector) Create(obj runtime.Object) {
	c.create(Create, obj)
}

// CreateOrUpdate creates the object if it does not exist, or updates it if it does.
func (c *NamespacedPatchCollector) CreateOrUpdate(obj runtime.Object) {
	c.create(CreateOrUpdate, obj)
}

// CreateIfNotExists creates the object only if it does not already exist.
func (c *NamespacedPatchCollector) CreateIfNotExists(obj runtime.Object) {
	c.create(CreateIfNotExists, obj)
}

func (c *NamespacedPatchCollector) create(operation CreateOperation, obj runtime.Object) {
	processed, err := utils.ToUnstructured(obj)
	if err != nil {
		c.collector.logger.Error("cannot convert data to unstructured object", log.Err(err))

		return
	}

	// Inject the fixed namespace before delegating to the underlying collector
	processed.SetNamespace(c.namespace)
	c.collector.createFromUnstructured(operation, processed)
}

// Delete removes the object using foreground cascading deletion.
func (c *NamespacedPatchCollector) Delete(apiVersion, kind, name string) {
	c.collector.delete(Delete, apiVersion, kind, c.namespace, name)
}

// DeleteInBackground removes the object immediately while the garbage collector
// deletes dependents in the background.
func (c *NamespacedPatchCollector) DeleteInBackground(apiVersion, kind, name string) {
	c.collector.delete(DeleteInBackground, apiVersion, kind, c.namespace, name)
}

// DeleteNonCascading removes the object without deleting its dependents (orphans them).
func (c *NamespacedPatchCollector) DeleteNonCascading(apiVersion, kind, name string) {
	c.collector.delete(DeleteNonCascading, apiVersion, kind, c.namespace, name)
}

// PatchWithJSON applies a RFC6902 JSON Patch to the object.
func (c *NamespacedPatchCollector) PatchWithJSON(jsonPatch any, apiVersion, kind, name string, opts ...pkg.PatchCollectorOption) {
	c.collector.patch(JSONPatch, jsonPatch, apiVersion, kind, c.namespace, name, opts...)
}

// PatchWithMerge applies a RFC7396 JSON Merge Patch to the object.
func (c *NamespacedPatchCollector) PatchWithMerge(mergePatch any, apiVersion, kind, name string, opts ...pkg.PatchCollectorOption) {
	c.collector.patch(MergePatch, mergePatch, apiVersion, kind, c.namespace, name, opts...)
}

// PatchWithJQ mutates the object using a jq filter expression.
func (c *NamespacedPatchCollector) PatchWithJQ(jqfilter, apiVersion, kind, name string, opts ...pkg.PatchCollectorOption) {
	c.collector.filter(jqfilter, apiVersion, kind, c.namespace, name, opts...)
}

// Operations returns all collected operations
func (c *NamespacedPatchCollector) Operations() []pkg.PatchCollectorOperation {
	return c.collector.Operations()
}

// WriteOutput serializes all collected operations as newline-delimited JSON.
func (c *NamespacedPatchCollector) WriteOutput(w io.Writer) error {
	return c.collector.WriteOutput(w)
}
