package kubernetes

import (
	"encoding/json"
	"io"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	service "github.com/deckhouse/module-sdk/pkg"
)

var _ service.PatchCollector = (*ObjectPatchCollector)(nil)

type ObjectPatchCollector struct {
	dataStorage []map[string]any
}

func NewObjectPatchCollector() *ObjectPatchCollector {
	return &ObjectPatchCollector{
		dataStorage: make([]map[string]any, 0),
	}
}

func (c *ObjectPatchCollector) collect(payload map[string]any) {
	c.dataStorage = append(c.dataStorage, payload)
}

func (c *ObjectPatchCollector) Create(data *unstructured.Unstructured) {
	c.create(Create, data)
}

func (c *ObjectPatchCollector) CreateOrUpdate(data *unstructured.Unstructured) {
	c.create(CreateOrUpdate, data)
}

func (c *ObjectPatchCollector) CreateIfNotExists(data *unstructured.Unstructured) {
	c.create(CreateIfNotExists, data)
}

func (c *ObjectPatchCollector) create(operation CreateOperation, obj *unstructured.Unstructured) {
	c.collect(map[string]any{"operation": operation, "object": obj})
}

func (c *ObjectPatchCollector) Delete(kind, namespace, name, apiVersion, subresource string) {
	c.delete(Delete, kind, namespace, name, apiVersion, subresource)
}

func (c *ObjectPatchCollector) DeleteInBackground(kind, namespace, name, apiVersion, subresource string) {
	c.delete(DeleteInBackground, kind, namespace, name, apiVersion, subresource)
}

func (c *ObjectPatchCollector) DeleteNonCascading(kind, namespace, name, apiVersion, subresource string) {
	c.delete(DeleteNonCascading, kind, namespace, name, apiVersion, subresource)
}

func (c *ObjectPatchCollector) delete(operation DeleteOperation, kind, namespace, name, apiVersion, subresource string) {
	ret := map[string]any{
		"operation": operation,
		"kind":      kind,
		"name":      name,
		"namespace": namespace,
	}

	if apiVersion != "" {
		ret["apiVersion"] = apiVersion
	}
	if subresource != "" {
		ret["subresource"] = subresource
	}

	c.collect(ret)
}

func (c *ObjectPatchCollector) JQPatch(kind, apiVersion, name, namespace string, filter string, subresource string) {
	c.patch(JQPatch, kind, apiVersion, name, namespace, filter, subresource, false)
}

func (c *ObjectPatchCollector) MergePatch(kind, apiVersion, name, namespace string, patch any, subresource string, ignoreMissingObjects bool) {
	c.patch(MergePatch, kind, apiVersion, name, namespace, patch, subresource, ignoreMissingObjects)
}

func (c *ObjectPatchCollector) JSONPatch(kind, apiVersion, name, namespace string, patch []any, subresource string, ignoreMissingObjects bool) {
	c.patch(JSONPatch, kind, apiVersion, name, namespace, patch, subresource, ignoreMissingObjects)
}

func (c *ObjectPatchCollector) patch(operation PatchOperation, kind, apiVersion, name, namespace string, patch any, subresource string, ignoreMissingObjects bool) {
	ret := map[string]any{
		"operation": operation,
		"kind":      kind,
		"name":      name,
		"namespace": namespace,
	}

	switch operation {
	case JQPatch:
		{
			ret["jqFilter"] = patch
		}
	case MergePatch:
		{
			ret["mergePatch"] = patch
		}
	case JSONPatch:
		{
			ret["jsonPatch"] = patch
		}
	}

	if apiVersion != "" {
		ret["apiVersion"] = apiVersion
	}
	if subresource != "" {
		ret["subresource"] = subresource
	}
	if ignoreMissingObjects {
		ret["ignoreMissingObjects"] = ignoreMissingObjects
	}

	c.collect(ret)
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
