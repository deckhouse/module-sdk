package registry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/pkg"
)

func TestRegister(t *testing.T) {
	t.Run("Hook with OnStartup and Kubernetes bindings should panic", func(t *testing.T) {
		hook := &pkg.HookConfig{
			OnStartup: &pkg.OrderedConfig{Order: 1},
			Kubernetes: []pkg.KubernetesConfig{
				{
					Name:       "test",
					APIVersion: "v1",
					Kind:       "Pod",
				},
			},
		}

		defer func() {
			r := recover()
			require.NotEmpty(t, r)
			assert.Equal(t, bindingsPanicMsg, r)
		}()

		RegisterFunc(hook, func(_ context.Context, _ *pkg.HookInput) error {
			return nil
		})
	})

	t.Run("Hook with OnStartup should not panic", func(t *testing.T) {
		hook := &pkg.HookConfig{
			OnStartup: &pkg.OrderedConfig{Order: 1},
		}

		defer func() {
			r := recover()
			assert.NotEqual(t, bindingsPanicMsg, r)
		}()

		RegisterFunc(hook, func(_ context.Context, _ *pkg.HookInput) error {
			return nil
		})
	})

	t.Run("Hook with Kubernetes binding should not panic", func(t *testing.T) {
		hook := &pkg.HookConfig{
			Kubernetes: []pkg.KubernetesConfig{
				{
					Name:       "test",
					APIVersion: "v1",
					Kind:       "Pod",
					// FilterFunc: nil,
				},
			},
		}

		defer func() {
			r := recover()
			assert.NotEqual(t, bindingsPanicMsg, r)
		}()

		RegisterFunc(hook, func(_ context.Context, _ *pkg.HookInput) error {
			return nil
		})
	})

	t.Run("Application hook with Kubernetes field should panic", func(t *testing.T) {
		hook := &pkg.HookConfig{
			Metadata: pkg.HookMetadata{
				Name: "test-hook",
				Path: "test/hooks",
			},
			HookType: pkg.HookTypeApplication,
			Kubernetes: []pkg.KubernetesConfig{
				{
					Name:       "test",
					APIVersion: "v1",
					Kind:       "Pod",
				},
			},
		}

		defer func() {
			r := recover()
			require.NotEmpty(t, r)
			assert.Contains(t, r.(string), "application hooks must use ApplicationKubernetes field")
		}()

		RegisterFunc(hook, func(_ context.Context, _ *pkg.ApplicationHookInput) error {
			return nil
		})
	})

	t.Run("Application hook with ApplicationKubernetes should not panic", func(t *testing.T) {
		hook := &pkg.HookConfig{
			Metadata: pkg.HookMetadata{
				Name: "test-hook",
				Path: "test/hooks",
			},
			HookType: pkg.HookTypeApplication,
			ApplicationKubernetes: []pkg.ApplicationKubernetesConfig{
				{
					Name:       "test",
					APIVersion: "v1",
					Kind:       "Pod",
				},
			},
		}

		defer func() {
			r := recover()
			assert.NotEqual(t, bindingsPanicMsg, r)
		}()

		RegisterFunc(hook, func(_ context.Context, _ *pkg.ApplicationHookInput) error {
			return nil
		})
	})
}
