package framework_test

import (
	"context"
	"testing"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/framework"
)

const node1YAML = `
---
apiVersion: v1
kind: Node
metadata:
  name: kube-worker-1
  labels:
    node-role: "testrole1"
`
const node2YAML = `
---
apiVersion: v1
kind: Node
metadata:
  name: kube-worker-2
  labels:
    node-role: "testrole1"
`

func Test_PrepareHookInput(t *testing.T) {
	config := &pkg.HookConfig{
		Metadata: pkg.HookMetadata{
			Name: "test-hook",
		},
		Kubernetes: []pkg.KubernetesConfig{
			{
				Name:       "nodes",
				APIVersion: "v1",
				Kind:       "Node",
				JqFilter:   ".metadata.name",
			},
		},
	}

	f := framework.NewHookFramework(t, config, func(ctx context.Context, input *pkg.HookInput) error {
		return nil
	})

	// Test with custom context and snapshots
	f.PrepareHookSnapshots(config, framework.InputSnapshots{
		"nodes": {
			node1YAML,
			node2YAML,
		},
	})

	// Test snapshots are correctly set
	if len(f.GetInput().Snapshots.Get("nodes")) != 2 {
		t.Errorf("Expected 2 node snapshots, got %d", len(f.GetInput().Snapshots.Get("nodes")))
	}

	// Test with multiple binding types
	multiConfig := &pkg.HookConfig{
		Metadata: pkg.HookMetadata{
			Name: "multi-test-hook",
		},
		Kubernetes: []pkg.KubernetesConfig{
			{
				Name:       "nodes",
				APIVersion: "v1",
				Kind:       "Node",
				JqFilter:   ".metadata.name",
			},
			{
				Name:       "node_roles",
				APIVersion: "v1",
				Kind:       "Node",
				JqFilter:   ".metadata.labels",
			},
		},
	}

	f = framework.NewHookFramework(t, multiConfig, func(ctx context.Context, input *pkg.HookInput) error {
		return nil
	})

	f.PrepareHookSnapshots(multiConfig, framework.InputSnapshots{
		"nodes": {
			node1YAML,
		},
		"node_roles": {
			node2YAML,
		},
	})

	// Verify collectors are initialized
	if f.GetInput().Values == nil || f.GetInput().ConfigValues == nil ||
		f.GetInput().PatchCollector == nil || f.GetInput().MetricsCollector == nil {
		t.Error("One or more collectors were not initialized")
	}

	// Verify logger is initialized
	if f.GetInput().Logger == nil {
		t.Error("Logger not initialized")
	}
}
