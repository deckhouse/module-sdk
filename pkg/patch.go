package pkg

import (
	"io"

	"github.com/deckhouse/module-sdk/pkg/utils"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type PatchCollector interface {
	Outputer

	Create(data *unstructured.Unstructured)
	CreateIfNotExists(data *unstructured.Unstructured)
	CreateOrUpdate(data *unstructured.Unstructured)
	Delete(kind string, namespace string, name string, apiVersion string, subresource string)
	DeleteInBackground(kind string, namespace string, name string, apiVersion string, subresource string)
	DeleteNonCascading(kind string, namespace string, name string, apiVersion string, subresource string)
	JQPatch(kind string, apiVersion string, name string, namespace string, filter string, subresource string, ignoreMissingObjects bool)
	MergePatch(kind string, apiVersion string, name string, namespace string, patch interface{}, subresource string, ignoreMissingObjects bool)
	JSONPatch(kind string, apiVersion string, name string, namespace string, patch []interface{}, subresource string, ignoreMissingObjects bool)
}

type PatchableValuesCollector interface {
	ArrayCount(path string) (int, error)
	Exists(path string) bool
	Get(path string) gjson.Result
	GetOk(path string) (gjson.Result, bool)
	GetPatches() []*utils.ValuesPatchOperation
	GetRaw(path string) interface{}
	Remove(path string)
	Set(path string, value interface{})
	WriteOutput(w io.Writer) error
}
