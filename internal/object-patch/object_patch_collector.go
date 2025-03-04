package objectpatch

import (
	"encoding/json"
	"io"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

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

func (c *ObjectPatchCollector) collect(payload Patch) {
	c.dataStorage = append(c.dataStorage, payload)
}

func (c *ObjectPatchCollector) Create(data any) {
	processed, err := utils.ToUnstructured(data)
	if err != nil {
		c.logger.Error("cannot convert data to unstructured object", log.Err(err))

		return
	}

	c.create(Create, processed)
}

func (c *ObjectPatchCollector) CreateOrUpdate(data any) {
	processed, err := utils.ToUnstructured(data)
	if err != nil {
		c.logger.Error("cannot convert data to unstructured object", log.Err(err))

		return
	}

	c.create(CreateOrUpdate, processed)
}

func (c *ObjectPatchCollector) CreateIfNotExists(data any) {
	processed, err := utils.ToUnstructured(data)
	if err != nil {
		c.logger.Error("cannot convert data to unstructured object", log.Err(err))

		return
	}

	c.create(CreateIfNotExists, processed)
}

func (c *ObjectPatchCollector) create(operation CreateOperation, obj *unstructured.Unstructured) {
	c.collect(Patch{"operation": operation, "object": obj})
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
	p := Patch{
		"operation":  operation,
		"apiVersion": apiVersion,
		"kind":       kind,
		"name":       name,
		"namespace":  namespace,
	}

	for _, opt := range opts {
		opt.Apply(p)
	}

	c.collect(p)
}

func (c *ObjectPatchCollector) JQPatch(filter string, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorPatchOption) {
	c.patch(JQPatch, filter, apiVersion, kind, namespace, name, opts...)
}

func (c *ObjectPatchCollector) MergePatch(patch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorPatchOption) {
	c.patch(MergePatch, patch, apiVersion, kind, namespace, name, opts...)
}

func (c *ObjectPatchCollector) JSONPatch(patch []any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorPatchOption) {
	c.patch(JSONPatch, patch, apiVersion, kind, namespace, name, opts...)
}

func (c *ObjectPatchCollector) patch(operation PatchOperation, patch any, apiVersion string, kind string, namespace string, name string, opts ...pkg.PatchCollectorPatchOption) {
	p := Patch{
		"operation":  operation,
		"apiVersion": apiVersion,
		"kind":       kind,
		"name":       name,
		"namespace":  namespace,
	}

	switch operation {
	case JQPatch:
		{
			p["jqFilter"] = patch
		}
	case MergePatch:
		{
			p["mergePatch"] = patch
		}
	case JSONPatch:
		{
			p["jsonPatch"] = patch
		}
	default:
		panic("not known operation")
	}

	for _, opt := range opts {
		opt.Apply(p)
	}

	c.collect(p)
}

// Operations returns all collected operations
func (c *ObjectPatchCollector) Operations() []Patch {
	return c.dataStorage
}

func (c *ObjectPatchCollector) WriteOutput(w io.Writer) error {
	for _, object := range c.dataStorage {
		err := json.NewEncoder(w).Encode(object)
		if err != nil {
			return err
		}
	}

	return nil
}
