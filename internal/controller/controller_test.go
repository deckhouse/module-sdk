package controller

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertAppKubernetesConfig(t *testing.T) {
	t.Run("minimal config copies required fields and sets queue", func(t *testing.T) {
		k := &pkg.ApplicationKubernetesConfig{
			Name:       "pods",
			APIVersion: "v1",
			Kind:       "Pod",
		}
		cfg := convertAppKubernetesConfig(k, "main")
		assert.Equal(t, "pods", cfg.Name)
		assert.Equal(t, "v1", cfg.APIVersion)
		assert.Equal(t, "Pod", cfg.Kind)
		assert.Equal(t, "main", cfg.Queue)
	})

	t.Run("NamespaceSelector is always nil for application hooks", func(t *testing.T) {
		k := &pkg.ApplicationKubernetesConfig{
			Name:       "pods",
			APIVersion: "v1",
			Kind:       "Pod",
		}
		cfg := convertAppKubernetesConfig(k, "main")
		assert.Nil(t, cfg.NamespaceSelector, "NamespaceSelector must be omitted for application hooks; addon-operator injects APPLICATION_NAMESPACE externally")
	})

	t.Run("NameSelector is converted when set", func(t *testing.T) {
		k := &pkg.ApplicationKubernetesConfig{
			Name:         "my-pod",
			APIVersion:   "v1",
			Kind:         "Pod",
			NameSelector: &pkg.NameSelector{MatchNames: []string{"pod-1", "pod-2"}},
		}
		cfg := convertAppKubernetesConfig(k, "main")
		require.NotNil(t, cfg.NameSelector)
		assert.Equal(t, []string{"pod-1", "pod-2"}, cfg.NameSelector.MatchNames)
	})

	t.Run("NameSelector is nil when not set", func(t *testing.T) {
		k := &pkg.ApplicationKubernetesConfig{
			Name:       "pods",
			APIVersion: "v1",
			Kind:       "Pod",
		}
		cfg := convertAppKubernetesConfig(k, "main")
		assert.Nil(t, cfg.NameSelector)
	})

	t.Run("FieldSelector is converted when set", func(t *testing.T) {
		k := &pkg.ApplicationKubernetesConfig{
			Name:       "pods",
			APIVersion: "v1",
			Kind:       "Pod",
			FieldSelector: &pkg.FieldSelector{
				MatchExpressions: []pkg.FieldSelectorRequirement{
					{Field: "status.phase", Operator: "In", Value: "Running"},
				},
			},
		}
		cfg := convertAppKubernetesConfig(k, "main")
		require.NotNil(t, cfg.FieldSelector)
		require.Len(t, cfg.FieldSelector.MatchExpressions, 1)
		assert.Equal(t, "status.phase", cfg.FieldSelector.MatchExpressions[0].Field)
		assert.Equal(t, "In", cfg.FieldSelector.MatchExpressions[0].Operator)
		assert.Equal(t, "Running", cfg.FieldSelector.MatchExpressions[0].Value)
	})

	t.Run("FieldSelector is nil when not set", func(t *testing.T) {
		k := &pkg.ApplicationKubernetesConfig{
			Name:       "pods",
			APIVersion: "v1",
			Kind:       "Pod",
		}
		cfg := convertAppKubernetesConfig(k, "main")
		assert.Nil(t, cfg.FieldSelector)
	})

	t.Run("KeepFullObjectsInMemory is true when JqFilter is empty", func(t *testing.T) {
		k := &pkg.ApplicationKubernetesConfig{
			Name:       "pods",
			APIVersion: "v1",
			Kind:       "Pod",
		}
		cfg := convertAppKubernetesConfig(k, "main")
		require.NotNil(t, cfg.KeepFullObjectsInMemory)
		assert.True(t, *cfg.KeepFullObjectsInMemory)
	})

	t.Run("KeepFullObjectsInMemory is false when JqFilter is set", func(t *testing.T) {
		k := &pkg.ApplicationKubernetesConfig{
			Name:     "pods",
			APIVersion: "v1",
			Kind:     "Pod",
			JqFilter: ".items[]",
		}
		cfg := convertAppKubernetesConfig(k, "main")
		require.NotNil(t, cfg.KeepFullObjectsInMemory)
		assert.False(t, *cfg.KeepFullObjectsInMemory)
	})

	t.Run("optional fields are passed through", func(t *testing.T) {
		executeHookOnEvents := false
		executeHookOnSync := true
		waitForSync := false
		allowFailure := true
		k := &pkg.ApplicationKubernetesConfig{
			Name:                         "pods",
			APIVersion:                   "v1",
			Kind:                         "Pod",
			LabelSelector:                &metav1.LabelSelector{MatchLabels: map[string]string{"app": "foo"}},
			ExecuteHookOnEvents:          &executeHookOnEvents,
			ExecuteHookOnSynchronization: &executeHookOnSync,
			WaitForSynchronization:       &waitForSync,
			AllowFailure:                 &allowFailure,
			ResynchronizationPeriod:      "15m",
		}
		cfg := convertAppKubernetesConfig(k, "queue-a")
		assert.Equal(t, "queue-a", cfg.Queue)
		require.NotNil(t, cfg.LabelSelector)
		assert.Equal(t, map[string]string{"app": "foo"}, cfg.LabelSelector.MatchLabels)
		require.NotNil(t, cfg.ExecuteHookOnEvents)
		assert.False(t, *cfg.ExecuteHookOnEvents)
		require.NotNil(t, cfg.ExecuteHookOnSynchronization)
		assert.True(t, *cfg.ExecuteHookOnSynchronization)
		require.NotNil(t, cfg.WaitForSynchronization)
		assert.False(t, *cfg.WaitForSynchronization)
		require.NotNil(t, cfg.AllowFailure)
		assert.True(t, *cfg.AllowFailure)
		assert.Equal(t, "15m", cfg.ResynchronizationPeriod)
	})

	t.Run("full config with all selectors still has NamespaceSelector nil", func(t *testing.T) {
		k := &pkg.ApplicationKubernetesConfig{
			Name:         "pods",
			APIVersion:   "v1",
			Kind:         "Pod",
			NameSelector: &pkg.NameSelector{MatchNames: []string{"x"}},
			FieldSelector: &pkg.FieldSelector{
				MatchExpressions: []pkg.FieldSelectorRequirement{
					{Field: "metadata.name", Operator: "Equals", Value: "y"},
				},
			},
		}
		cfg := convertAppKubernetesConfig(k, "main")
		require.NotNil(t, cfg.NameSelector)
		require.NotNil(t, cfg.FieldSelector)
		assert.Nil(t, cfg.NamespaceSelector, "application hook config must never set NamespaceSelector")
	})
}
