package main_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/framework"

	singlefileexample "singlefileexample"
)

// hookConfig mirrors the configuration in main.go but is built locally so
// that the functional test does not depend on the global registry being
// pristine when go test runs the package.
var hookConfig = &pkg.HookConfig{
	Kubernetes: []pkg.KubernetesConfig{
		{
			Name:       singlefileexample.SnapshotKey,
			APIVersion: "v1",
			Kind:       "Pod",
			NamespaceSelector: &pkg.NamespaceSelector{
				NameSelector: &pkg.NameSelector{MatchNames: []string{"kube-system"}},
			},
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"component": "kube-apiserver"},
			},
			JqFilter: ".metadata.name",
		},
	},
}

// TestHandle_DiscoversAPIServerPodsFromCluster is the canonical
// end-to-end test for this example: kube-apiserver Pods are seeded into
// the fake cluster, the framework filters them by namespace + labels,
// passes the names to the hook, and the hook writes them to internal
// values.
func TestHandle_DiscoversAPIServerPodsFromCluster(t *testing.T) {
	const state = `
---
apiVersion: v1
kind: Pod
metadata:
  name: kube-apiserver-1
  namespace: kube-system
  labels:
    component: kube-apiserver
---
apiVersion: v1
kind: Pod
metadata:
  name: kube-apiserver-2
  namespace: kube-system
  labels:
    component: kube-apiserver
---
apiVersion: v1
kind: Pod
metadata:
  name: not-an-apiserver
  namespace: kube-system
  labels:
    component: scheduler
---
apiVersion: v1
kind: Pod
metadata:
  name: kube-apiserver-other
  namespace: default
  labels:
    component: kube-apiserver
`

	f := framework.HookExecutionConfigInit(t,
		hookConfig,
		singlefileexample.Handle,
		`{}`, `{}`,
	)
	f.KubeStateSet(state)
	f.RunHook()

	require.NoError(t, f.HookError())

	// Only the two kube-system + apiserver Pods should be in the snapshot.
	snaps := f.Snapshots().Get(singlefileexample.SnapshotKey)
	require.Len(t, snaps, 2)

	got := f.ValuesGet("test.internal.apiServers")
	require.True(t, got.Exists())

	arr := got.Array()
	names := make([]string, 0, len(arr))
	for _, item := range arr {
		names = append(names, item.String())
	}
	assert.ElementsMatch(t, []string{"kube-apiserver-1", "kube-apiserver-2"}, names)
}

// TestHandle_NoPodsResultsInEmptyValues verifies the empty path: when
// the cluster has nothing matching the binding, the hook still writes an
// empty list to keep the rest of the chart deterministic.
func TestHandle_NoPodsResultsInEmptyValues(t *testing.T) {
	f := framework.HookExecutionConfigInit(t,
		hookConfig,
		singlefileexample.Handle,
		`{}`, `{}`,
	)
	f.RunHook()

	require.NoError(t, f.HookError())

	got := f.ValuesGet("test.internal.apiServers")
	require.True(t, got.Exists())
	assert.Empty(t, got.Array())
}
