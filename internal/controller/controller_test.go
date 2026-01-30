package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/internal/executor"
	"github.com/deckhouse/module-sdk/pkg"
)

type mockExecutor struct {
	isAppHook bool
	config    pkg.Config
}

func (m *mockExecutor) Config() pkg.Config {
	return m.config
}

func (m *mockExecutor) Execute(_ context.Context, _ executor.Request) (executor.Result, error) {
	return nil, nil
}

func (m *mockExecutor) IsApplicationHook() bool {
	return m.isAppHook
}

// Module Hook without the namespaceSelector.
// Waiting: The NamespaceSelector remains nil (monitors the entire cluster or works by default).
func Test_remapHookConfigToHookConfig_ModuleHook_PreservesNilSelector(t *testing.T) {
	t.Setenv(pkg.EnvApplicationNamespace, "some-app-ns")

	config := &pkg.HookConfig{
		Metadata: pkg.HookMetadata{Name: "module-hook-global"},
		Kubernetes: []pkg.KubernetesConfig{
			{Name: "nodes", APIVersion: "v1", Kind: "Node"},
		},
	}

	mockExec := &mockExecutor{isAppHook: false, config: config}

	result, err := remapHookConfigToHookConfig(mockExec.Config())
	require.NoError(t, err)

	require.Len(t, result.Kubernetes, 1)
	assert.Nil(t, result.Kubernetes[0].NamespaceSelector)
}

// Module Hook with an explicitly specified namespace.
// Waiting: The configuration is saved as it is.
func Test_remapHookConfigToHookConfig_ModuleHook_PreservesCustomSelector(t *testing.T) {
	targetNs := "kube-system"

	config := &pkg.HookConfig{
		Metadata: pkg.HookMetadata{Name: "module-hook-system"},
		Kubernetes: []pkg.KubernetesConfig{
			{
				Name:       "pods",
				APIVersion: "v1",
				Kind:       "Pod",
				NamespaceSelector: &pkg.NamespaceSelector{
					NameSelector: &pkg.NameSelector{
						MatchNames: []string{targetNs},
					},
				},
			},
		},
	}

	mockExec := &mockExecutor{isAppHook: false, config: config}

	result, err := remapHookConfigToHookConfig(mockExec.Config())
	require.NoError(t, err)

	require.Len(t, result.Kubernetes, 1)
	assert.NotNil(t, result.Kubernetes[0].NamespaceSelector)

	assert.Equal(t, []string{targetNs}, result.Kubernetes[0].NamespaceSelector.NameSelector.MatchNames)
}
