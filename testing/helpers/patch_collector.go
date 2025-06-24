package helpers

import (
	"testing"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/mock"
	"github.com/stretchr/testify/assert"
)

func PreparePatchCollector(t *testing.T, patches ...any) pkg.OutputPatchCollector {
	pc := mock.NewPatchCollectorMock(t)

	for _, patch := range patches {
		switch p := patch.(type) {
		case Create:
			pc.CreateMock.Set(func(obj any) {
				assert.Equal(t, p.Object, obj, "Create object mismatch")
			})
		case CreateIfNotExists:
			pc.CreateIfNotExistsMock.Set(func(obj any) {
				assert.Equal(t, p.Object, obj, "CreateIfNotExists object mismatch")
			})
		case CreateOrUpdate:
			pc.CreateOrUpdateMock.Set(func(obj any) {
				assert.Equal(t, p.Object, obj, "CreateOrUpdate object mismatch")
			})
		case Delete:
			pc.DeleteMock.Set(func(apiVersion, kind, namespace, name string) {
				assert.Equal(t, p.ApiVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case DeleteInBackground:
			pc.DeleteInBackgroundMock.Set(func(apiVersion, kind, namespace, name string) {
				assert.Equal(t, p.ApiVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case DeleteNonCascading:
			pc.DeleteNonCascadingMock.Set(func(apiVersion, kind, namespace, name string) {
				assert.Equal(t, p.ApiVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case JSONPatch:
			pc.JSONPatchMock.Set(func(jsonPatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				assert.Equal(t, p.JsonPatch, jsonPatch, "JSON patch mismatch")
				assert.Equal(t, p.ApiVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case MergePatch:
			pc.MergePatchMock.Set(func(mergePatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				assert.Equal(t, p.MergePatch, mergePatch, "Merge patch mismatch")
				assert.Equal(t, p.ApiVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case JQFilter:
			pc.JQFilterMock.Set(func(jqFilter string, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				assert.Equal(t, p.JQFilter, jqFilter, "JQ filter mismatch")
				assert.Equal(t, p.ApiVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case PatchWithJQ:
			pc.PatchWithJQMock.Set(func(jqfilter, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				assert.Equal(t, p.JQFilter, jqfilter, "JQ filter mismatch")
				assert.Equal(t, p.ApiVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case PatchWithJSON:
			pc.PatchWithJSONMock.Set(func(jsonPatch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				assert.Equal(t, p.JsonPatch, jsonPatch, "JSON patch mismatch")
				assert.Equal(t, p.ApiVersion, apiVersion, "API version mismatch")
				assert.Equal(t, p.Kind, kind, "Kind mismatch")
				assert.Equal(t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(t, p.Name, name, "Name mismatch")
			})
		case PatchWithMerge:
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

// Patch type definitions for different patching operations
type Create struct {
	Object any
}

type CreateIfNotExists struct {
	Object any
}

type CreateOrUpdate struct {
	Object any
}

type Delete struct {
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
}

type DeleteInBackground struct {
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
}

type DeleteNonCascading struct {
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
}

type JSONPatch struct {
	JsonPatch  any
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
	Options    []pkg.PatchCollectorOption
}

type MergePatch struct {
	MergePatch any
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
	Options    []pkg.PatchCollectorOption
}

type JQFilter struct {
	JQFilter   string
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
	Options    []pkg.PatchCollectorOption
}

type PatchWithJQ struct {
	JQFilter   string
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
}

type PatchWithJSON struct {
	JsonPatch  any
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
	Options    []pkg.PatchCollectorOption
}

type PatchWithMerge struct {
	MergePatch any
	ApiVersion string
	Kind       string
	Namespace  string
	Name       string
	Options    []pkg.PatchCollectorOption
}
