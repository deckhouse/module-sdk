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

func (c *ObjectPatchCollector) Create(obj any, opts ...pkg.PatchCollectorCreateOption) {
	c.create(Create, obj, opts...)
}

func (c *ObjectPatchCollector) CreateOrUpdate(obj any, opts ...pkg.PatchCollectorCreateOption) {
	c.create(CreateOrUpdate, obj, opts...)
}

func (c *ObjectPatchCollector) CreateIfNotExists(obj any, opts ...pkg.PatchCollectorCreateOption) {
	c.create(CreateIfNotExists, obj, opts...)
}

func (c *ObjectPatchCollector) create(operation CreateOperation, obj any, opts ...pkg.PatchCollectorCreateOption) {
	processed, err := utils.ToUnstructured(obj)
	if err != nil {
		c.logger.Error("cannot convert data to unstructured object", log.Err(err))

		return
	}

	p := &Patch{
		patchValues: map[string]any{
			"operation": operation,
			"object":    processed,
		},
	}

	for _, opt := range opts {
		opt.Apply(p)
	}

	c.collect(p)
}

func (c *ObjectPatchCollector) Delete(apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorDeleteOption) {
	c.delete(Delete, apiVersion, kind, namespace, name, opts...)
}

func (c *ObjectPatchCollector) DeleteInBackground(apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorDeleteOption) {
	c.delete(DeleteInBackground, apiVersion, kind, namespace, name, opts...)
}

func (c *ObjectPatchCollector) DeleteNonCascading(apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorDeleteOption) {
	c.delete(DeleteNonCascading, apiVersion, kind, namespace, name, opts...)
}

func (c *ObjectPatchCollector) delete(operation DeleteOperation, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorDeleteOption) {
	p := &Patch{
		patchValues: map[string]any{
			"operation":  operation,
			"apiVersion": apiVersion,
			"kind":       kind,
			"name":       name,
			"namespace":  namespace,
		},
	}

	for _, opt := range opts {
		opt.Apply(p)
	}

	c.collect(p)
}

func (c *ObjectPatchCollector) MergePatch(patch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorPatchOption) {
	c.patch(MergePatch, patch, apiVersion, kind, namespace, name, opts...)
}

func (c *ObjectPatchCollector) JSONPatch(patch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorPatchOption) {
	c.patch(JSONPatch, patch, apiVersion, kind, namespace, name, opts...)
}

func (c *ObjectPatchCollector) patch(operation PatchOperation, patch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorPatchOption) {
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

func (c *ObjectPatchCollector) JQFilter(filter string, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorFilterOption) {
	c.filter(filter, apiVersion, kind, namespace, name, opts...)
}

func (c *ObjectPatchCollector) filter(patch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorFilterOption) {
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
