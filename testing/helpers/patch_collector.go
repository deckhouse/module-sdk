// nolint: revive
package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/mock"
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
				assert.Equal(t, p.APIVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case *deleteInBackground:
			pc.DeleteInBackgroundMock.Set(func(apiVersion, kind, namespace, name string) {
				assert.Equal(t, p.APIVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case *deleteNonCascading:
			pc.DeleteNonCascadingMock.Set(func(apiVersion, kind, namespace, name string) {
				assert.Equal(t, p.APIVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case *jsonPatch:
			pc.JSONPatchMock.Set(func(jsonPatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				assert.Equal(t, p.JSONPatch, jsonPatch, "JSON patch mismatch")
				assert.Equal(t, p.APIVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case *mergePatch:
			pc.MergePatchMock.Set(func(mergePatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				assert.Equal(t, p.MergePatch, mergePatch, "Merge patch mismatch")
				assert.Equal(t, p.APIVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case *jqFilter:
			pc.JQFilterMock.Set(func(jqFilter string, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				assert.Equal(t, p.JQFilter, jqFilter, "JQ filter mismatch")
				assert.Equal(t, p.APIVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case *patchWithJQ:
			pc.PatchWithJQMock.Set(func(jqfilter, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				assert.Equal(t, p.JQFilter, jqfilter, "JQ filter mismatch")
				assert.Equal(t, p.APIVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case *patchWithJSON:
			pc.PatchWithJSONMock.Set(func(jsonPatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				assert.Equal(t, p.JSONPatch, jsonPatch, "JSON patch mismatch")
				assert.Equal(t, p.APIVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case *patchWithMerge:
			pc.PatchWithMergeMock.Set(func(mergePatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				assert.Equal(t, p.MergePatch, mergePatch, "Merge patch mismatch")
				assert.Equal(t, p.APIVersion, apiVersion, "API version mismatch")
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
	return &typedPatch{wrapped: &delete{APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type delete struct {
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}

func NewDeleteInBackground(apiVersion, kind, namespace, name string) *typedPatch {
	return &typedPatch{wrapped: &deleteInBackground{APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type deleteInBackground struct {
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}

func NewDeleteNonCascading(apiVersion, kind, namespace, name string) *typedPatch {
	return &typedPatch{wrapped: &deleteNonCascading{APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type deleteNonCascading struct {
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}

func NewJSONPatch(jsonPatchRaw any, apiVersion, kind, namespace, name string) *typedPatch {
	return &typedPatch{wrapped: &jsonPatch{JSONPatch: jsonPatchRaw, APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type jsonPatch struct {
	JSONPatch  any
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}

func NewMergePatch(mergePatchRaw any, apiVersion, kind, namespace, name string) *typedPatch {
	return &typedPatch{wrapped: &mergePatch{MergePatch: mergePatchRaw, APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type mergePatch struct {
	MergePatch any
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}

func NewJQFilter(jqFilterStr, apiVersion, kind, namespace, name string) *typedPatch {
	return &typedPatch{wrapped: &jqFilter{JQFilter: jqFilterStr, APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type jqFilter struct {
	JQFilter   string
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}

func NewPatchWithJQ(jqFilter, apiVersion, kind, namespace, name string) *typedPatch {
	return &typedPatch{wrapped: &patchWithJQ{JQFilter: jqFilter, APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type patchWithJQ struct {
	JQFilter   string
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}

func NewPatchWithJSON(jsonPatch any, apiVersion, kind, namespace, name string) *typedPatch {
	return &typedPatch{wrapped: &patchWithJSON{JSONPatch: jsonPatch, APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type patchWithJSON struct {
	JSONPatch  any
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}

func NewPatchWithMerge(mergePatch any, apiVersion, kind, namespace, name string) *typedPatch {
	return &typedPatch{wrapped: &patchWithMerge{MergePatch: mergePatch, APIVersion: apiVersion, Kind: kind, Namespace: namespace, Name: name}}
}

type patchWithMerge struct {
	MergePatch any
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}
