package hooks

import (
	"context"
	"testing"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestApplicationHookRegistration runs from a path containing ".../hooks/001-hook-one/",
// so the registry's extractHookMetadata() can extract Name and Path from the call stack.
func TestApplicationHookRegistration(t *testing.T) {
	t.Run("Application hook with Kubernetes binding should not panic", func(t *testing.T) {
		hook := &pkg.ApplicationHookConfig{
			Kubernetes: []pkg.ApplicationKubernetesConfig{
				{
					Name:       "test",
					APIVersion: "v1",
					Kind:       "Pod",
				},
			},
		}

		defer func() {
			r := recover()
			assert.NotEqual(t, registry.BindingsPanicMsg, r)
		}()

		registry.RegisterFunc(hook, func(_ context.Context, _ *pkg.ApplicationHookInput) error {
			return nil
		})
	})

	t.Run("Application hook metadata is extracted from call stack", func(t *testing.T) {
		hook := &pkg.ApplicationHookConfig{
			Kubernetes: []pkg.ApplicationKubernetesConfig{
				{
					Name:       "test",
					APIVersion: "v1",
					Kind:       "Pod",
				},
			},
		}

		registry.RegisterFunc(hook, func(_ context.Context, _ *pkg.ApplicationHookInput) error {
			return nil
		})

		appHooks := registry.Registry().ApplicationHooks()
		require.GreaterOrEqual(t, len(appHooks), 1, "at least one application hook should be registered")
		registered := appHooks[len(appHooks)-1]
		assert.NotEmpty(t, registered.Config.Metadata.Name, "Metadata.Name should be set by extractHookMetadata from call stack")
		assert.NotEmpty(t, registered.Config.Metadata.Path, "Metadata.Path should be set by extractHookMetadata from call stack")
	})
}
