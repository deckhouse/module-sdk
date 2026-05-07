package framework_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deckhouse/module-sdk/pkg"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	"github.com/deckhouse/module-sdk/testing/framework"
)

// nodeBindingConfig is a sample hook config used across tests.
var nodeBindingConfig = &pkg.HookConfig{
	Metadata: pkg.HookMetadata{Name: "test-hook"},
	Kubernetes: []pkg.KubernetesConfig{
		{
			Name:       "nodes",
			APIVersion: "v1",
			Kind:       "Node",
			JqFilter:   `{name: .metadata.name, role: .metadata.labels["node-role"]}`,
		},
	},
}

const initialNodes = `
---
apiVersion: v1
kind: Node
metadata:
  name: kube-worker-1
  labels:
    node-role: worker
---
apiVersion: v1
kind: Node
metadata:
  name: kube-worker-2
  labels:
    node-role: worker
---
apiVersion: v1
kind: Node
metadata:
  name: kube-master-1
  labels:
    node-role: master
`

// TestSnapshotsFromKubeState exercises the deckhouse-style flow:
// KubeStateSet → RunHook → assert on snapshots / patches / values.
func TestSnapshotsFromKubeState(t *testing.T) {
	type filteredNode struct {
		Name string `json:"name"`
		Role string `json:"role"`
	}

	handler := func(_ context.Context, input *pkg.HookInput) error {
		nodes, err := objectpatch.UnmarshalToStruct[filteredNode](input.Snapshots, "nodes")
		if err != nil {
			return err
		}
		names := make([]string, 0, len(nodes))
		for _, n := range nodes {
			names = append(names, n.Name)
		}
		input.Values.Set("module.nodes.count", len(nodes))
		input.Values.Set("module.nodes.names", names)
		return nil
	}

	hec := framework.HookExecutionConfigInit(t, nodeBindingConfig, handler, `{}`, `{}`)

	hec.KubeStateSet(initialNodes)
	hec.RunHook()

	require.NoError(t, hec.HookError())

	got := hec.Snapshots().Get("nodes")
	assert.Len(t, got, 3)

	assert.Equal(t, int64(3), hec.ValuesGet("module.nodes.count").Int())

	names := hec.ValuesGet("module.nodes.names").Array()
	assert.Len(t, names, 3)
}

// TestKubeStateTransitions verifies that KubeStateSet can be called multiple
// times and each RunHook sees the latest state.
func TestKubeStateTransitions(t *testing.T) {
	handler := func(_ context.Context, input *pkg.HookInput) error {
		input.Values.Set("count", len(input.Snapshots.Get("nodes")))
		return nil
	}

	hec := framework.HookExecutionConfigInit(t, nodeBindingConfig, handler, `{}`, `{}`)

	hec.KubeStateSet(initialNodes)
	hec.RunHook()
	require.NoError(t, hec.HookError())
	assert.Equal(t, int64(3), hec.ValuesGet("count").Int())

	hec.KubeStateSet(`
---
apiVersion: v1
kind: Node
metadata:
  name: kube-worker-1
`)
	hec.RunHook()
	require.NoError(t, hec.HookError())
	assert.Equal(t, int64(1), hec.ValuesGet("count").Int())
}

// TestPatchCollectorAppliesToFakeCluster verifies that Create/Delete/Patch
// operations performed by the hook are replayed against the fake cluster.
func TestPatchCollectorAppliesToFakeCluster(t *testing.T) {
	cfg := &pkg.HookConfig{
		Metadata:   pkg.HookMetadata{Name: "patch-hook"},
		Kubernetes: nil,
	}

	handler := func(_ context.Context, input *pkg.HookInput) error {
		// Create a configmap.
		cm := &corev1.ConfigMap{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
			ObjectMeta: metav1.ObjectMeta{Name: "demo", Namespace: "default"},
			Data:       map[string]string{"hello": "world"},
		}
		input.PatchCollector.Create(cm)

		// Issue a merge patch (target may not exist yet, so use IgnoreMissingObject).
		input.PatchCollector.PatchWithMerge(
			map[string]any{"data": map[string]string{"hello": "patched"}},
			"v1", "ConfigMap", "default", "demo",
		)
		return nil
	}

	hec := framework.HookExecutionConfigInit(t, cfg, handler, `{}`, `{}`)
	hec.RunHook()
	require.NoError(t, hec.HookError())

	// Two patch operations recorded.
	ops := hec.PatchedOperations()
	require.Len(t, ops, 2)
	assert.Equal(t, framework.PatchTypeCreate, ops[0].Type)
	assert.Equal(t, framework.PatchTypeMergePatch, ops[1].Type)

	// The fake cluster should now contain the configmap with patched value.
	cm := hec.KubernetesResource("ConfigMap", "default", "demo")
	require.NotNil(t, cm)

	data := nestedMap(cm.Object, "data")
	assert.Equal(t, "patched", data["hello"])
}

// TestPatchTypesAreApplied exercises Delete and JSONPatch operations.
func TestPatchTypesAreApplied(t *testing.T) {
	cfg := &pkg.HookConfig{Metadata: pkg.HookMetadata{Name: "complex"}}

	handler := func(_ context.Context, input *pkg.HookInput) error {
		input.PatchCollector.Create(&corev1.ConfigMap{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
			ObjectMeta: metav1.ObjectMeta{Name: "to-delete", Namespace: "kube-system"},
		})
		input.PatchCollector.Delete("v1", "ConfigMap", "kube-system", "to-delete")

		// Create a config map and JSON-patch a key.
		input.PatchCollector.Create(&corev1.ConfigMap{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
			ObjectMeta: metav1.ObjectMeta{Name: "json-patched", Namespace: "default"},
			Data:       map[string]string{"a": "1"},
		})
		input.PatchCollector.PatchWithJSON(
			[]map[string]any{
				{"op": "add", "path": "/data/b", "value": "2"},
			},
			"v1", "ConfigMap", "default", "json-patched",
		)
		return nil
	}

	hec := framework.HookExecutionConfigInit(t, cfg, handler, `{}`, `{}`)
	hec.RunHook()
	require.NoError(t, hec.HookError())

	// Deleted resource should be gone.
	assert.Nil(t, hec.KubernetesResource("ConfigMap", "kube-system", "to-delete"))

	// JSON-patched resource has both keys.
	cm := hec.KubernetesResource("ConfigMap", "default", "json-patched")
	require.NotNil(t, cm)
	data := nestedMap(cm.Object, "data")
	assert.Equal(t, "1", data["a"])
	assert.Equal(t, "2", data["b"])
}

// TestValuesAndConfigValuesArePatched ensures values written by the hook
// (via input.Values.Set) are visible after RunHook.
func TestValuesAndConfigValuesArePatched(t *testing.T) {
	cfg := &pkg.HookConfig{Metadata: pkg.HookMetadata{Name: "values-hook"}}

	handler := func(_ context.Context, input *pkg.HookInput) error {
		input.Values.Set("module.replicaCount", 3)
		input.Values.Set("module.feature.enabled", true)
		input.ConfigValues.Set("module.profile", "prod")
		return nil
	}

	initial := `
module:
  replicaCount: 1
  feature:
    enabled: false
`
	hec := framework.HookExecutionConfigInit(t, cfg, handler, initial, `{}`)
	hec.RunHook()
	require.NoError(t, hec.HookError())

	assert.Equal(t, int64(3), hec.ValuesGet("module.replicaCount").Int())
	assert.True(t, hec.ValuesGet("module.feature.enabled").Bool())
	assert.Equal(t, "prod", hec.ConfigValuesGet("module.profile").String())
}

// TestNamespaceSelectorBindings verifies snapshot generation for a binding
// scoped to specific namespaces.
func TestNamespaceSelectorBindings(t *testing.T) {
	cfg := &pkg.HookConfig{
		Metadata: pkg.HookMetadata{Name: "ns-hook"},
		Kubernetes: []pkg.KubernetesConfig{
			{
				Name:       "system_pods",
				APIVersion: "v1",
				Kind:       "Pod",
				NamespaceSelector: &pkg.NamespaceSelector{
					NameSelector: &pkg.NameSelector{MatchNames: []string{"kube-system"}},
				},
				JqFilter: `.metadata.name`,
			},
		},
	}

	handler := func(_ context.Context, input *pkg.HookInput) error {
		input.Values.Set("count", len(input.Snapshots.Get("system_pods")))
		return nil
	}

	hec := framework.HookExecutionConfigInit(t, cfg, handler, `{}`, `{}`)
	hec.KubeStateSet(`
---
apiVersion: v1
kind: Pod
metadata:
  name: kube-proxy
  namespace: kube-system
---
apiVersion: v1
kind: Pod
metadata:
  name: my-app
  namespace: default
`)
	hec.RunHook()
	require.NoError(t, hec.HookError())

	assert.Equal(t, int64(1), hec.ValuesGet("count").Int())
}

// TestLabelSelectorBinding verifies snapshot filtering by label selector.
func TestLabelSelectorBinding(t *testing.T) {
	cfg := &pkg.HookConfig{
		Metadata: pkg.HookMetadata{Name: "label-hook"},
		Kubernetes: []pkg.KubernetesConfig{
			{
				Name:       "workers",
				APIVersion: "v1",
				Kind:       "Node",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"node-role": "worker"},
				},
				JqFilter: `.metadata.name`,
			},
		},
	}
	handler := func(_ context.Context, input *pkg.HookInput) error {
		input.Values.Set("workerCount", len(input.Snapshots.Get("workers")))
		return nil
	}

	hec := framework.HookExecutionConfigInit(t, cfg, handler, `{}`, `{}`)
	hec.KubeStateSet(initialNodes)
	hec.RunHook()
	require.NoError(t, hec.HookError())
	assert.Equal(t, int64(2), hec.ValuesGet("workerCount").Int())
}

// TestHookErrorIsCaptured verifies HookError reflects the handler's return value.
func TestHookErrorIsCaptured(t *testing.T) {
	cfg := &pkg.HookConfig{Metadata: pkg.HookMetadata{Name: "error-hook"}}
	handler := func(_ context.Context, _ *pkg.HookInput) error {
		return assertErr("boom")
	}
	hec := framework.HookExecutionConfigInit(t, cfg, handler, `{}`, `{}`)
	hec.RunHook()
	require.Error(t, hec.HookError())
	assert.Contains(t, hec.HookError().Error(), "boom")
}

// TestRegisterCRD allows resources of an unknown kind to be used in state YAML.
func TestRegisterCRD(t *testing.T) {
	cfg := &pkg.HookConfig{
		Metadata: pkg.HookMetadata{Name: "crd-hook"},
		Kubernetes: []pkg.KubernetesConfig{
			{
				Name:       "widgets",
				APIVersion: "example.com/v1alpha1",
				Kind:       "Widget",
				JqFilter:   `.metadata.name`,
			},
		},
	}
	handler := func(_ context.Context, input *pkg.HookInput) error {
		input.Values.Set("widgets", len(input.Snapshots.Get("widgets")))
		return nil
	}

	hec := framework.NewHookExecutionConfig(t, cfg, handler,
		framework.WithCRD("example.com", "v1alpha1", "Widget", true),
	)

	hec.KubeStateSet(`
---
apiVersion: example.com/v1alpha1
kind: Widget
metadata:
  name: widget-a
  namespace: default
spec:
  size: 10
---
apiVersion: example.com/v1alpha1
kind: Widget
metadata:
  name: widget-b
  namespace: default
`)
	hec.RunHook()
	require.NoError(t, hec.HookError())
	assert.Equal(t, int64(2), hec.ValuesGet("widgets").Int())

	// Lookup individual CR via KubernetesResource.
	w := hec.KubernetesResource("Widget", "default", "widget-a")
	require.NotNil(t, w)
	size, _ := nestedInt(w.Object, "spec", "size")
	assert.Equal(t, int64(10), size)
}

// TestLoggerOutputCaptured verifies the logger output is captured.
func TestLoggerOutputCaptured(t *testing.T) {
	cfg := &pkg.HookConfig{Metadata: pkg.HookMetadata{Name: "log-hook"}}
	handler := func(_ context.Context, input *pkg.HookInput) error {
		input.Logger.Info("hello from hook")
		return nil
	}
	hec := framework.HookExecutionConfigInit(t, cfg, handler, `{}`, `{}`)
	hec.RunHook()
	require.NoError(t, hec.HookError())
	assert.True(t, strings.Contains(hec.LoggerOutput().String(), "hello from hook"))
}

// === helpers ===

type assertErr string

func (a assertErr) Error() string { return string(a) }

func nestedMap(obj map[string]any, fields ...string) map[string]string {
	v := any(obj)
	for _, f := range fields {
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = m[f]
	}
	out := map[string]string{}
	if m, ok := v.(map[string]any); ok {
		for k, val := range m {
			s, _ := val.(string)
			out[k] = s
		}
		return out
	}
	return nil
}

func nestedInt(obj map[string]any, fields ...string) (int64, bool) {
	v := any(obj)
	for _, f := range fields {
		m, ok := v.(map[string]any)
		if !ok {
			return 0, false
		}
		v = m[f]
	}
	switch x := v.(type) {
	case int64:
		return x, true
	case float64:
		return int64(x), true
	case int:
		return int64(x), true
	}
	return 0, false
}
