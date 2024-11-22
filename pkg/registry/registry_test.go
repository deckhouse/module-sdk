package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/pkg"
)

func TestRegister(t *testing.T) {
	t.Run("Hook with OnStartup and Kubernetes bindings should panic", func(t *testing.T) {
		hook := &pkg.Hook{
			Config: &pkg.HookConfig{
				OnStartup: &pkg.OrderedConfig{Order: 1},
				Kubernetes: []pkg.KubernetesConfig{
					{
						Name:       "test",
						APIVersion: "v1",
						Kind:       "Pod",
						// FilterFunc: nil,
					},
				},
			},
			ReconcileFunc: nil,
		}

		defer func() {
			r := recover()
			require.NotEmpty(t, r)
			assert.Equal(t, bindingsPanicMsg, r)
		}()
		Registry().Add(hook)
	})

	t.Run("Hook with OnStartup should not panic", func(t *testing.T) {
		hook := &pkg.Hook{
			Config: &pkg.HookConfig{
				OnStartup: &pkg.OrderedConfig{Order: 1},
			},
			ReconcileFunc: nil,
		}

		defer func() {
			r := recover()
			assert.NotEqual(t, bindingsPanicMsg, r)
		}()
		Registry().Add(hook)
	})

	t.Run("Hook with Kubernetes binding should not panic", func(t *testing.T) {
		hook := &pkg.Hook{
			Config: &pkg.HookConfig{
				Kubernetes: []pkg.KubernetesConfig{
					{
						Name:       "test",
						APIVersion: "v1",
						Kind:       "Pod",
						// FilterFunc: nil,
					},
				},
			},
			ReconcileFunc: nil,
		}

		defer func() {
			r := recover()
			assert.NotEqual(t, bindingsPanicMsg, r)
		}()
		Registry().Add(hook)
	})
}
