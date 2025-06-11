package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/deckhouse/deckhouse/pkg/log"
	objectpatch "github.com/deckhouse/module-sdk/internal/object-patch"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/jq"
	"github.com/deckhouse/module-sdk/testing/mock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
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
			fmt.Println("JSON Snapshot:", string(res.String()))

			formattedSnapshots[snapBindingName] = append(formattedSnapshots[snapBindingName], objectpatch.Snapshot(res.String()))
		}
	}

	f.HookInput.Snapshots = formattedSnapshots
}
