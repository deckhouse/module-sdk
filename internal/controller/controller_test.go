package controller

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/internal/executor"
	"github.com/deckhouse/module-sdk/pkg"
)

type mockExecutor struct {
	isAppHook bool
	config    *pkg.HookConfig
}

func (m *mockExecutor) Config() *pkg.HookConfig {
	if m.isAppHook {
		m.config.HookType = pkg.HookTypeApplication
	} else {
		m.config.HookType = pkg.HookTypeModule
	}
	return m.config
}

func (m *mockExecutor) Execute(_ context.Context, _ executor.Request) (executor.Result, error) {
	return nil, nil
}

func (m *mockExecutor) IsApplicationHook() bool {
	return m.isAppHook
}

// Application Hook without namespace selector.
// Waiting: Namespace is automatically injected from the env variable.
func Test_remapHookConfigToHookConfig_ApplicationHook_InjectsNamespace(t *testing.T) {
	appName := "my-test-app"
	t.Setenv(pkg.EnvApplicationNamespace, appName)

	config := &pkg.HookConfig{
		Metadata: pkg.HookMetadata{Name: "app-hook-simple"},
		HookType: pkg.HookTypeApplication,
		Kubernetes: []pkg.KubernetesConfig{
			{Name: "pods", APIVersion: "v1", Kind: "Pod"},
		},
	}

	mockExec := &mockExecutor{isAppHook: true, config: config}

	result, err := remapHookConfigToHookConfig(mockExec.Config())
	require.NoError(t, err)

	require.Len(t, result.Kubernetes, 1)
	assert.NotNil(t, result.Kubernetes[0].NamespaceSelector)
	assert.NotNil(t, result.Kubernetes[0].NamespaceSelector.NameSelector)
	assert.Equal(t, []string{appName}, result.Kubernetes[0].NamespaceSelector.NameSelector.MatchNames)
}

// Case 3: Application Hook, but forgot to set the environment variable.
// Waiting: The function returns an error (Fail Fast), the config is not generated.
func Test_remapHookConfigToHookConfig_ApplicationHook_ErrorsOnMissingEnv(t *testing.T) {
	os.Unsetenv(pkg.EnvApplicationNamespace)

	config := &pkg.HookConfig{
		Metadata: pkg.HookMetadata{Name: "app-hook-broken"},
		HookType: pkg.HookTypeApplication,
		Kubernetes: []pkg.KubernetesConfig{
			{Name: "pods", APIVersion: "v1", Kind: "Pod"},
		},
	}

	mockExec := &mockExecutor{isAppHook: true, config: config}

	result, err := remapHookConfigToHookConfig(mockExec.Config())

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "application hook \"app-hook-broken\" requires APPLICATION_NAMESPACE env var to be set")
}

// Module Hook without the namespaceSelector.
// Waiting: The NamespaceSelector remains nil (monitors the entire cluster or works by default).
func Test_remapHookConfigToHookConfig_ModuleHook_PreservesNilSelector(t *testing.T) {
	t.Setenv(pkg.EnvApplicationNamespace, "some-app-ns")

	config := &pkg.HookConfig{
		Metadata: pkg.HookMetadata{Name: "module-hook-global"},
		HookType: pkg.HookTypeModule,
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
		HookType: pkg.HookTypeModule,
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
