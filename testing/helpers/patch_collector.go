package helpers

import (
	"testing"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/mock"
	"github.com/stretchr/testify/assert"
)

type typedPatch struct {
	wrapped any
}

func PreparePatchCollector(t *testing.T, patches ...*typedPatch) pkg.OutputPatchCollector {
	pc := mock.NewPatchCollectorMock(t)

	for _, patch := range patches {
		switch p := patch.wrapped.(type) {
		case *create:
			pc.CreateMock.Set(func(obj any) {
				assert.Equal(t, p.Object, obj, "Create object mismatch")
			})
		case *createIfNotExists:
			pc.CreateIfNotExistsMock.Set(func(obj any) {
				assert.Equal(t, p.Object, obj, "CreateIfNotExists object mismatch")
			})
		case *createOrUpdate:
			pc.CreateOrUpdateMock.Set(func(obj any) {
				assert.Equal(t, p.Object, obj, "CreateOrUpdate object mismatch")
			})
		case *delete:
			pc.DeleteMock.Set(func(apiVersion, kind, namespace, name string) {
				assert.Equal(t, p.ApiVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case *deleteInBackground:
			pc.DeleteInBackgroundMock.Set(func(apiVersion, kind, namespace, name string) {
				assert.Equal(t, p.ApiVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case *deleteNonCascading:
			pc.DeleteNonCascadingMock.Set(func(apiVersion, kind, namespace, name string) {
				assert.Equal(t, p.ApiVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case *jsonPatch:
			pc.JSONPatchMock.Set(func(jsonPatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				assert.Equal(t, p.JsonPatch, jsonPatch, "JSON patch mismatch")
				assert.Equal(t, p.ApiVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case *mergePatch:
			pc.MergePatchMock.Set(func(mergePatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				assert.Equal(t, p.MergePatch, mergePatch, "Merge patch mismatch")
				assert.Equal(t, p.ApiVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case *jqFilter:
			pc.JQFilterMock.Set(func(jqFilter string, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				assert.Equal(t, p.JQFilter, jqFilter, "JQ filter mismatch")
				assert.Equal(t, p.ApiVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case *patchWithJQ:
			pc.PatchWithJQMock.Set(func(jqfilter, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				assert.Equal(t, p.JQFilter, jqfilter, "JQ filter mismatch")
				assert.Equal(t, p.ApiVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case *patchWithJSON:
			pc.PatchWithJSONMock.Set(func(jsonPatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				assert.Equal(t, p.JsonPatch, jsonPatch, "JSON patch mismatch")
				assert.Equal(t, p.ApiVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case *patchWithMerge:
			pc.PatchWithMergeMock.Set(func(mergePatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				assert.Equal(t, p.MergePatch, mergePatch, "Merge patch mismatch")
				assert.Equal(t, p.ApiVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		default:
			t.Fatalf("Unsupported patch type: %T", p)
		}
	}

	return pc
}
func NewCreate(obj any) *typedPatch {
	return &typedPatch{wrapped: &create{Object: obj}}
}

type create struct {
	Object any
}

func NewCreateIfNotExists(obj any) *typedPatch {
	return &typedPatch{wrapped: &createIfNotExists{Object: obj}}
}

type createIfNotExists struct {
	Object any
}

func NewCreateOrUpdate(obj any) *typedPatch {
	return &typedPatch{wrapped: &createOrUpdate{Object: obj}}
}

type createOrUpdate struct {
	Object any
}

func NewDelete(apiVersion, kind, namespace, name string) *typedPatch {
	return &typedPatch{wrapped: &delete{ApiVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type delete struct {
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
}

func NewDeleteInBackground(apiVersion, kind, namespace, name string) *typedPatch {
	return &typedPatch{wrapped: &deleteInBackground{ApiVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type deleteInBackground struct {
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
}

func NewDeleteNonCascading(apiVersion, kind, namespace, name string) *typedPatch {
	return &typedPatch{wrapped: &deleteNonCascading{ApiVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type deleteNonCascading struct {
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
}

func NewJSONPatch(JSONPatch any, apiVersion, kind, namespace, name string) *typedPatch {
	return &typedPatch{wrapped: &jsonPatch{JsonPatch: JSONPatch, ApiVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type jsonPatch struct {
	JsonPatch  any
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
}

func NewMergePatch(MergePatch any, apiVersion, kind, namespace, name string) *typedPatch {
	return &typedPatch{wrapped: &mergePatch{MergePatch: MergePatch, ApiVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type mergePatch struct {
	MergePatch any
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
}

func NewJQFilter(JQFilter, apiVersion, kind, namespace, name string) *typedPatch {
	return &typedPatch{wrapped: &jqFilter{JQFilter: JQFilter, ApiVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type jqFilter struct {
	JQFilter   string
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
}

func NewPatchWithJQ(jqFilter, apiVersion, kind, namespace, name string) *typedPatch {
	return &typedPatch{wrapped: &patchWithJQ{JQFilter: jqFilter, ApiVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type patchWithJQ struct {
	JQFilter   string
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
}

func NewPatchWithJSON(jsonPatch any, apiVersion, kind, namespace, name string) *typedPatch {
	return &typedPatch{wrapped: &patchWithJSON{JsonPatch: jsonPatch, ApiVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type patchWithJSON struct {
	JsonPatch  any
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
}

func NewPatchWithMerge(mergePatch any, apiVersion, kind, namespace, name string) *typedPatch {
	return &typedPatch{wrapped: &patchWithMerge{MergePatch: mergePatch, ApiVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type patchWithMerge struct {
	MergePatch any
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
}
