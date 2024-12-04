package hookinfolder

import (
	"context"

	"github.com/deckhouse/module-sdk/pkg"
	metrics "github.com/deckhouse/module-sdk/pkg/metric/operation"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

var _ = registry.RegisterFunc(configMetricsCollector, handlerHookMetricsCollector)

var configMetricsCollector = &pkg.HookConfig{
	OnBeforeHelm: &pkg.OrderedConfig{Order: 1},
	Kubernetes: []pkg.KubernetesConfig{
		{
			Name:       nodeInfoSnapshotName,
			APIVersion: "v1",
			Kind:       "Node",
			JqFilter:   applyNodeJQFilter,
		},
	},
}

func handlerHookMetricsCollector(_ context.Context, input *pkg.HookInput) error {
	input.Logger.Info("hello from metrics hook")

	input.MetricsCollector.Add("stub-add-metric", 1, map[string]string{"node_found": "node_name"}, metrics.WithGroup("my-group"))

	input.MetricsCollector.Set("stub-set-metric", 1, map[string]string{"node_found": "node_name"}, metrics.WithGroup("my-group"))

	input.MetricsCollector.Inc("stub-inc-metric", map[string]string{"node_found": "node_name"}, metrics.WithGroup("my-group"))

	return nil
}
