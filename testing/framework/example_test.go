package framework_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deckhouse/module-sdk/pkg"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	"github.com/deckhouse/module-sdk/testing/framework"
)

// TestExample_DeckhouseStyle is a comprehensive end-to-end example mirroring
// how a deckhouse hook test is typically written. It is intentionally verbose
// to serve as documentation.
//
// The hook under test counts the number of running pods in the "default"
// namespace, writes the count to values, and creates a status ConfigMap.
func TestExample_DeckhouseStyle(t *testing.T) {
	// 1. Hook config — same as in production code.
	cfg := &pkg.HookConfig{
		Metadata: pkg.HookMetadata{Name: "pod-counter"},
		Kubernetes: []pkg.KubernetesConfig{
			{
				Name:       "pods",
				APIVersion: "v1",
				Kind:       "Pod",
				NamespaceSelector: &pkg.NamespaceSelector{
					NameSelector: &pkg.NameSelector{MatchNames: []string{"default"}},
				},
				JqFilter: `{name: .metadata.name, phase: .status.phase}`,
			},
		},
	}

	type podSnap struct {
		Name  string `json:"name"`
		Phase string `json:"phase"`
	}

	// 2. Hook handler — also same as in production code.
	handler := func(_ context.Context, input *pkg.HookInput) error {
		pods, err := objectpatch.UnmarshalToStruct[podSnap](input.Snapshots, "pods")
		if err != nil {
			return err
		}

		var running int
		for _, p := range pods {
			if p.Phase == "Running" {
				running++
			}
		}

		input.Values.Set("podCounter.running", running)

		input.PatchCollector.Create(&corev1.ConfigMap{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
			ObjectMeta: metav1.ObjectMeta{Name: "pod-counter-status", Namespace: "default"},
			Data:       map[string]string{"running": fmt.Sprintf("%d", running)},
		})
		return nil
	}

	// 3. Initialise the framework as in deckhouse: HookExecutionConfigInit.
	hec := framework.HookExecutionConfigInit(t, cfg, handler, `{}`, `{}`)

	// 4. Describe the cluster state with YAML.
	hec.KubeStateSet(`
---
apiVersion: v1
kind: Pod
metadata:
  name: app-1
  namespace: default
status:
  phase: Running
---
apiVersion: v1
kind: Pod
metadata:
  name: app-2
  namespace: default
status:
  phase: Pending
---
apiVersion: v1
kind: Pod
metadata:
  name: kube-proxy
  namespace: kube-system
status:
  phase: Running
`)

	// 5. Run the hook.
	hec.RunHook()

	// 6. Inspect the results.
	require.NoError(t, hec.HookError())

	// Snapshots respect the namespace selector.
	require.Len(t, hec.Snapshots().Get("pods"), 2)

	// Values produced by the hook.
	assert.Equal(t, int64(1), hec.ValuesGet("podCounter.running").Int())

	// Patch operations recorded.
	ops := hec.PatchedOperations()
	require.Len(t, ops, 1)
	assert.Equal(t, framework.PatchTypeCreate, ops[0].Type)

	// And the create operation has actually been applied to the fake cluster:
	cm := hec.KubernetesResource("ConfigMap", "default", "pod-counter-status")
	require.NotNil(t, cm)
	data, _ := cm.Object["data"].(map[string]any)
	require.NotNil(t, data)
	assert.Equal(t, "1", data["running"])
}
