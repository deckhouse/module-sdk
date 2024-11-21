package pkg

import (
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/deckhouse/module-sdk/pkg/utils"
)

type PatchCollector interface {
	Outputer

	Create(data *unstructured.Unstructured)
	CreateIfNotExists(data *unstructured.Unstructured)
	CreateOrUpdate(data *unstructured.Unstructured)

	Delete(kind string, namespace string, name string, apiVersion string, subresource string)
	DeleteInBackground(kind string, namespace string, name string, apiVersion string, subresource string)
	DeleteNonCascading(kind string, namespace string, name string, apiVersion string, subresource string)

	JQPatch(kind string, apiVersion string, name string, namespace string, filter string, subresource string)
	MergePatch(kind string, apiVersion string, name string, namespace string, patch any, subresource string, ignoreMissingObjects bool)
	JSONPatch(kind string, apiVersion string, name string, namespace string, patch []any, subresource string, ignoreMissingObjects bool)
}

type PatchableValuesCollector interface {
	Outputer

	ArrayCount(path string) (int, error)
	Exists(path string) bool
	Get(path string) gjson.Result
	GetOk(path string) (gjson.Result, bool)
	GetPatches() []*utils.ValuesPatchOperation
	GetRaw(path string) any
	Remove(path string)
	Set(path string, value any)
}
