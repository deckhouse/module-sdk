package objectpatch

import (
	"encoding/json"
	"io"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/utils"
)

// Compile-time interface compliance check
var _ pkg.PatchCollector = (*PatchCollector)(nil)

// PatchCollector collects Kubernetes object patch operations to be
// applied after hook execution. Supports create, delete, and various patch
// operations (JSON Patch, Merge Patch, JQ filter).
// Note: This collector is not thread-safe; do not use concurrently.
type PatchCollector struct {
	dataStorage []Patch
	logger      *log.Logger
}

// NewCollector creates an empty collector ready to accumulate patch operations.
func NewCollector(logger *log.Logger) *PatchCollector {
	return &PatchCollector{
		dataStorage: make([]Patch, 0),
		logger:      logger,
	}
}

func (c *PatchCollector) collect(payload *Patch) {
	if payload == nil {
		return
	}

	c.dataStorage = append(c.dataStorage, *payload)
}

func (c *PatchCollector) Create(obj any) {
	c.create(Create, obj)
}

func (c *PatchCollector) CreateOrUpdate(obj any) {
	c.create(CreateOrUpdate, obj)
}

func (c *PatchCollector) CreateIfNotExists(obj any) {
	c.create(CreateIfNotExists, obj)
}

func (c *PatchCollector) create(operation CreateOperation, obj any) {
	processed, err := utils.ToUnstructured(obj)
	if err != nil {
		c.logger.Error("cannot convert data to unstructured object", log.Err(err))

		return
	}

	c.createFromUnstructured(operation, processed)
}

// createFromUnstructured collects a create operation with a pre-converted object.
// Used internally and by NamespacedPatchCollector for namespace injection.
func (c *PatchCollector) createFromUnstructured(operation CreateOperation, obj any) {
	p := &Patch{
		patchValues: map[string]any{
			"operation": operation,
			"object":    obj,
		},
	}

	c.collect(p)
}

func (c *PatchCollector) Delete(apiVersion string, kind string, namespace string, name string) {
	c.delete(Delete, apiVersion, kind, namespace, name)
}

func (c *PatchCollector) DeleteInBackground(apiVersion string, kind string, namespace string, name string) {
	c.delete(DeleteInBackground, apiVersion, kind, namespace, name)
}

func (c *PatchCollector) DeleteNonCascading(apiVersion string, kind string, namespace string, name string) {
	c.delete(DeleteNonCascading, apiVersion, kind, namespace, name)
}

func (c *PatchCollector) delete(operation DeleteOperation, apiVersion string, kind string, namespace string, name string) {
	p := &Patch{
		patchValues: map[string]any{
			"operation":  operation,
			"apiVersion": apiVersion,
			"kind":       kind,
			"name":       name,
			"namespace":  namespace,
		},
	}

	c.collect(p)
}

func (c *PatchCollector) MergePatch(patch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorOption) {
	c.patch(MergePatch, patch, apiVersion, kind, namespace, name, opts...)
}

func (c *PatchCollector) JSONPatch(patch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorOption) {
	c.patch(JSONPatch, patch, apiVersion, kind, namespace, name, opts...)
}

func (c *PatchCollector) PatchWithJSON(jsonPatch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorOption) {
	c.patch(JSONPatch, jsonPatch, apiVersion, kind, namespace, name, opts...)
}

func (c *PatchCollector) PatchWithMerge(mergePatch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorOption) {
	c.patch(MergePatch, mergePatch, apiVersion, kind, namespace, name, opts...)
}

func (c *PatchCollector) PatchWithJQ(jqfilter string, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorOption) {
	c.filter(jqfilter, apiVersion, kind, namespace, name, opts...)
}

func (c *PatchCollector) patch(operation PatchOperation, patch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorOption) {
	p := &Patch{
		patchValues: map[string]any{
			"operation":  operation,
			"apiVersion": apiVersion,
			"kind":       kind,
			"name":       name,
			"namespace":  namespace,
		},
	}

	switch operation {
	case JQPatch:
		panic("filter jq operation in patch method")
	case MergePatch:
		p.patchValues["mergePatch"] = patch
	case JSONPatch:
		p.patchValues["jsonPatch"] = patch
	default:
		panic("not known operation")
	}

	for _, opt := range opts {
		opt.Apply(p)
	}

	c.collect(p)
}

func (c *PatchCollector) JQFilter(filter string, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorOption) {
	c.filter(filter, apiVersion, kind, namespace, name, opts...)
}

func (c *PatchCollector) filter(patch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorOption) {
	p := &Patch{
		patchValues: map[string]any{
			"operation":  JQPatch,
			"apiVersion": apiVersion,
			"kind":       kind,
			"name":       name,
			"namespace":  namespace,
			"jqFilter":   patch,
		},
	}

	for _, opt := range opts {
		opt.Apply(p)
	}

	c.collect(p)
}

// Operations returns all collected operations
func (c *PatchCollector) Operations() []pkg.PatchCollectorOperation {
	operations := make([]pkg.PatchCollectorOperation, 0, len(c.dataStorage))

	for _, object := range c.dataStorage {
		operations = append(operations, &object)
	}

	return operations
}

// WriteOutput serializes all collected operations as newline-delimited JSON.
func (c *PatchCollector) WriteOutput(w io.Writer) error {
	for _, object := range c.dataStorage {
		err := json.NewEncoder(w).Encode(object.patchValues)
		if err != nil {
			return err
		}
	}

	return nil
}
