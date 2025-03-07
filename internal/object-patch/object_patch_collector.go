package objectpatch

import (
	"encoding/json"
	"io"

	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/utils"
)

var _ pkg.PatchCollector = (*ObjectPatchCollector)(nil)

type ObjectPatchCollector struct {
	dataStorage []Patch
	logger      *log.Logger
}

func NewObjectPatchCollector(logger *log.Logger) *ObjectPatchCollector {
	return &ObjectPatchCollector{
		dataStorage: make([]Patch, 0),
		logger:      logger,
	}
}

func (c *ObjectPatchCollector) collect(payload *Patch) {
	if payload == nil {
		return
	}

	c.dataStorage = append(c.dataStorage, *payload)
}

func (c *ObjectPatchCollector) Create(obj any) {
	c.create(Create, obj)
}

func (c *ObjectPatchCollector) CreateOrUpdate(obj any) {
	c.create(CreateOrUpdate, obj)
}

func (c *ObjectPatchCollector) CreateIfNotExists(obj any) {
	c.create(CreateIfNotExists, obj)
}

func (c *ObjectPatchCollector) create(operation CreateOperation, obj any) {
	processed, err := utils.ToUnstructured(obj)
	if err != nil {
		c.logger.Error("cannot convert data to unstructured object", log.Err(err))

		return
	}

	p := &Patch{
		kind: operationCreate,
		patchValues: map[string]any{
			"operation": operation,
			"object":    processed,
		},
	}

	c.collect(p)
}

func (c *ObjectPatchCollector) Delete(apiVersion string, kind string, namespace string, name string) {
	c.delete(Delete, apiVersion, kind, namespace, name)
}

func (c *ObjectPatchCollector) DeleteInBackground(apiVersion string, kind string, namespace string, name string) {
	c.delete(DeleteInBackground, apiVersion, kind, namespace, name)
}

func (c *ObjectPatchCollector) DeleteNonCascading(apiVersion string, kind string, namespace string, name string) {
	c.delete(DeleteNonCascading, apiVersion, kind, namespace, name)
}

func (c *ObjectPatchCollector) delete(operation DeleteOperation, apiVersion string, kind string, namespace string, name string) {
	p := &Patch{
		kind: operationDelete,
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

func (c *ObjectPatchCollector) MergePatch(patch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorOption) {
	c.patch(MergePatch, patch, apiVersion, kind, namespace, name, opts...)
}

func (c *ObjectPatchCollector) JSONPatch(patch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorOption) {
	c.patch(JSONPatch, patch, apiVersion, kind, namespace, name, opts...)
}

func (c *ObjectPatchCollector) PatchWithJSON(jsonPatch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorOption) {
	c.patch(JSONPatch, jsonPatch, apiVersion, kind, namespace, name, opts...)
}

func (c *ObjectPatchCollector) PatchWithMerge(mergePatch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorOption) {
	c.patch(MergePatch, mergePatch, apiVersion, kind, namespace, name, opts...)
}

func (c *ObjectPatchCollector) PatchWithJQ(jqfilter string, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorOption) {
	c.filter(jqfilter, apiVersion, kind, namespace, name, opts...)
}

func (c *ObjectPatchCollector) patch(operation PatchOperation, patch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorOption) {
	p := &Patch{
		kind: operationPatch,
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
		{
			p.patchValues["mergePatch"] = patch
		}
	case JSONPatch:
		{
			p.patchValues["jsonPatch"] = patch
		}
	default:
		panic("not known operation")
	}

	for _, opt := range opts {
		opt.Apply(p)
	}

	c.collect(p)
}

func (c *ObjectPatchCollector) JQFilter(filter string, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorOption) {
	c.filter(filter, apiVersion, kind, namespace, name, opts...)
}

func (c *ObjectPatchCollector) filter(patch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorOption) {
	p := &Patch{
		kind: operationFilter,
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
func (c *ObjectPatchCollector) Operations() []pkg.PatchCollectorOperation {
	operations := make([]pkg.PatchCollectorOperation, 0, len(c.dataStorage))

	for _, object := range c.dataStorage {
		operations = append(operations, &object)
	}

	return operations
}

func (c *ObjectPatchCollector) WriteOutput(w io.Writer) error {
	for _, object := range c.dataStorage {
		err := json.NewEncoder(w).Encode(object.patchValues)
		if err != nil {
			return err
		}
	}

	return nil
}
