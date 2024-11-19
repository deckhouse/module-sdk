package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gohook "github.com/deckhouse/module-sdk/pkg/hook"
	gohookcfg "github.com/deckhouse/module-sdk/pkg/hook/config"
)

func TestRegister(t *testing.T) {
	t.Run("Hook with OnStartup and Kubernetes bindings should panic", func(t *testing.T) {
		hook := gohook.NewGoHook(
			&gohookcfg.HookConfig{
				OnStartup: 1,
				Kubernetes: []gohookcfg.KubernetesConfig{
					{
						Name:       "test",
						ApiVersion: "v1",
						Kind:       "Pod",
						// FilterFunc: nil,
					},
				},
			},
			nil,
		)

		defer func() {
			r := recover()
			require.NotEmpty(t, r)
			assert.Equal(t, bindingsPanicMsg, r)
		}()
		Registry().Add(hook)
	})

	t.Run("Hook with OnStartup should not panic", func(t *testing.T) {
		hook := gohook.NewGoHook(
			&gohookcfg.HookConfig{
				OnStartup: 1,
			},
			nil,
		)

		defer func() {
			r := recover()
			assert.NotEqual(t, bindingsPanicMsg, r)
		}()
		Registry().Add(hook)
	})

	t.Run("Hook with Kubernetes binding should not panic", func(t *testing.T) {
		hook := gohook.NewGoHook(
			&gohookcfg.HookConfig{
				Kubernetes: []gohookcfg.KubernetesConfig{
					{
						Name:       "test",
						ApiVersion: "v1",
						Kind:       "Pod",
						// FilterFunc: nil,
					},
				},
			},
			nil,
		)

		defer func() {
			r := recover()
			assert.NotEqual(t, bindingsPanicMsg, r)
		}()
		Registry().Add(hook)
	})
}
