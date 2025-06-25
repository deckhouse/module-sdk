package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/deckhouse/deckhouse/pkg/log"

	objectpatch "github.com/deckhouse/module-sdk/internal/object-patch"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/jq"
	"github.com/deckhouse/module-sdk/testing/mock"
)

type HookFramework struct {
	t             *testing.T
	Config        *pkg.HookConfig
	ReconcileFunc pkg.ReconcileFunc
	HookInput     *pkg.HookInput
}

func NewHookFramework(t *testing.T, config *pkg.HookConfig, f pkg.ReconcileFunc) *HookFramework {
	return &HookFramework{
		t:             t,
		Config:        config,
		ReconcileFunc: f,
		HookInput: &pkg.HookInput{
			Snapshots:        mock.NewSnapshotsMock(t),
			Values:           mock.NewPatchableValuesCollectorMock(t),
			ConfigValues:     mock.NewPatchableValuesCollectorMock(t),
			PatchCollector:   mock.NewPatchCollectorMock(t),
			MetricsCollector: mock.NewMetricsCollectorMock(t),
			DC:               mock.NewDependencyContainerMock(t),
			Logger:           log.NewNop(),
		},
	}
}

func (f *HookFramework) GetInput() *pkg.HookInput {
	if f.HookInput == nil {
		f.t.Fatal("HookInput is not initialized")
	}
	return f.HookInput
}

type InputSnapshots map[string][]string

func (f *HookFramework) PrepareHookSnapshots(config *pkg.HookConfig, inputSnapshots InputSnapshots) {
	formattedSnapshots := make(objectpatch.Snapshots, len(inputSnapshots))
	for snapBindingName, snaps := range inputSnapshots {
		var (
			err   error
			query *jq.Query
		)

		for _, v := range config.Kubernetes {
			if v.Name == snapBindingName {
				fmt.Println("Using JQ filter:", v.JqFilter)
				query, err = jq.NewQuery(v.JqFilter)
				assert.NoError(f.t, err, "Failed to create JQ query from filter: %s", v.JqFilter)
			}
		}

		for _, snap := range snaps {
			var yml map[string]interface{}

			err := yaml.Unmarshal([]byte(snap), &yml)
			assert.NoError(f.t, err, "Failed to unmarshal snapshot YAML: %s", snap)

			jsonSnap, err := json.Marshal(yml)
			assert.NoError(f.t, err, "Failed to marshal snapshot to JSON: %s", snap)

			fmt.Println("JSON Snapshot:", string(jsonSnap))

			res, err := query.FilterStringObject(context.TODO(), string(jsonSnap))
			assert.NoError(f.t, err, "Failed to filter snapshot with JQ query: %s", jsonSnap)
			fmt.Println("JSON Snapshot:", res.String())

			formattedSnapshots[snapBindingName] = append(formattedSnapshots[snapBindingName], objectpatch.Snapshot(res.String()))
		}
	}

	f.HookInput.Snapshots = formattedSnapshots
}

func (f *HookFramework) PreparePatchCollector(patches ...any) {
	pc := mock.NewPatchCollectorMock(f.t)

	for _, patch := range patches {
		switch p := patch.(type) {
		case Create:
			pc.CreateMock.Set(func(obj any) {
				assert.Equal(f.t, p.Object, obj, "Create object mismatch")
			})
		case CreateIfNotExists:
			pc.CreateIfNotExistsMock.Set(func(obj any) {
				assert.Equal(f.t, p.Object, obj, "CreateIfNotExists object mismatch")
			})
		case CreateOrUpdate:
			pc.CreateOrUpdateMock.Set(func(obj any) {
				assert.Equal(f.t, p.Object, obj, "CreateOrUpdate object mismatch")
			})
		case Delete:
			pc.DeleteMock.Set(func(apiVersion, kind, namespace, name string) {
				assert.Equal(f.t, p.APIVersion, apiVersion, "API version mismatch")
				assert.Equal(f.t, p.Kind, kind, "Kind mismatch")
				assert.Equal(f.t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(f.t, p.Name, name, "Name mismatch")
			})
		case DeleteInBackground:
			pc.DeleteInBackgroundMock.Set(func(apiVersion, kind, namespace, name string) {
				assert.Equal(f.t, p.APIVersion, apiVersion, "API version mismatch")
				assert.Equal(f.t, p.Kind, kind, "Kind mismatch")
				assert.Equal(f.t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(f.t, p.Name, name, "Name mismatch")
			})
		case DeleteNonCascading:
			pc.DeleteNonCascadingMock.Set(func(apiVersion, kind, namespace, name string) {
				assert.Equal(f.t, p.APIVersion, apiVersion, "API version mismatch")
				assert.Equal(f.t, p.Kind, kind, "Kind mismatch")
				assert.Equal(f.t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(f.t, p.Name, name, "Name mismatch")
			})
		case JSONPatch:
			pc.JSONPatchMock.Set(func(jsonPatch any, apiVersion, kind, namespace, name string, _ ...pkg.PatchCollectorOption) {
				assert.Equal(f.t, p.JSONPatch, jsonPatch, "JSON patch mismatch")
				assert.Equal(f.t, p.APIVersion, apiVersion, "API version mismatch")
				assert.Equal(f.t, p.Kind, kind, "Kind mismatch")
				assert.Equal(f.t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(f.t, p.Name, name, "Name mismatch")
			})
		case MergePatch:
			pc.MergePatchMock.Set(func(mergePatch any, apiVersion, kind, namespace, name string, _ ...pkg.PatchCollectorOption) {
				assert.Equal(f.t, p.MergePatch, mergePatch, "Merge patch mismatch")
				assert.Equal(f.t, p.APIVersion, apiVersion, "API version mismatch")
				assert.Equal(f.t, p.Kind, kind, "Kind mismatch")
				assert.Equal(f.t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(f.t, p.Name, name, "Name mismatch")
			})
		case JQFilter:
			pc.JQFilterMock.Set(func(jqFilter string, apiVersion, kind, namespace, name string, _ ...pkg.PatchCollectorOption) {
				assert.Equal(f.t, p.JQFilter, jqFilter, "JQ filter mismatch")
				assert.Equal(f.t, p.APIVersion, apiVersion, "API version mismatch")
				assert.Equal(f.t, p.Kind, kind, "Kind mismatch")
				assert.Equal(f.t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(f.t, p.Name, name, "Name mismatch")
			})
		case PatchWithJQ:
			pc.PatchWithJQMock.Set(func(jqfilter, apiVersion, kind, namespace, name string, _ ...pkg.PatchCollectorOption) {
				assert.Equal(f.t, p.JQFilter, jqfilter, "JQ filter mismatch")
				assert.Equal(f.t, p.APIVersion, apiVersion, "API version mismatch")
				assert.Equal(f.t, p.Kind, kind, "Kind mismatch")
				assert.Equal(f.t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(f.t, p.Name, name, "Name mismatch")
			})
		case PatchWithJSON:
			pc.PatchWithJSONMock.Set(func(jsonPatch any, apiVersion, kind, namespace, name string, _ ...pkg.PatchCollectorOption) {
				assert.Equal(f.t, p.JSONPatch, jsonPatch, "JSON patch mismatch")
				assert.Equal(f.t, p.APIVersion, apiVersion, "API version mismatch")
				assert.Equal(f.t, p.Kind, kind, "Kind mismatch")
				assert.Equal(f.t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(f.t, p.Name, name, "Name mismatch")
			})
		case PatchWithMerge:
			pc.PatchWithMergeMock.Set(func(mergePatch any, apiVersion, kind, namespace, name string, _ ...pkg.PatchCollectorOption) {
				assert.Equal(f.t, p.MergePatch, mergePatch, "Merge patch mismatch")
				assert.Equal(f.t, p.APIVersion, apiVersion, "API version mismatch")
				assert.Equal(f.t, p.Kind, kind, "Kind mismatch")
				assert.Equal(f.t, p.Namespace, namespace, "Namespace mismatch")
				assert.Equal(f.t, p.Name, name, "Name mismatch")
			})
		default:
			f.t.Fatalf("Unsupported patch type: %T", p)
		}
	}

	f.HookInput.PatchCollector = pc
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
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}

type DeleteInBackground struct {
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}

type DeleteNonCascading struct {
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}

type JSONPatch struct {
	JSONPatch  any
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
	Options    []pkg.PatchCollectorOption
}

type MergePatch struct {
	MergePatch any
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
	Options    []pkg.PatchCollectorOption
}

type JQFilter struct {
	JQFilter   string
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
	Options    []pkg.PatchCollectorOption
}

type PatchWithJQ struct {
	JQFilter   string
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}

type PatchWithJSON struct {
	JSONPatch  any
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
	Options    []pkg.PatchCollectorOption
}

type PatchWithMerge struct {
	MergePatch any
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
	Options    []pkg.PatchCollectorOption
}

func (f *HookFramework) Execute(ctx context.Context) error {
	err := f.ReconcileFunc(ctx, f.HookInput)
	if err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	return nil
}
