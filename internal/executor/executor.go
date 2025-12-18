package executor

import (
	"context"

	bindingcontext "github.com/deckhouse/module-sdk/internal/binding-context"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/utils"
)

type Executor interface {
	Config() *pkg.HookConfig
	Execute(ctx context.Context, req Request) (Result, error)
}

type Request interface {
	GetValues() (map[string]any, error)
	GetConfigValues() (map[string]any, error)
	GetBindingContexts() ([]bindingcontext.BindingContext, error)
	GetDependencyContainer() pkg.DependencyContainer
}

type Result interface {
	ValuesPatches() map[utils.ValuesPatchType]pkg.OutputPatchableValuesCollector
	ObjectPatchCollector() pkg.OutputPatchCollector
	MetricsCollector() pkg.OutputMetricsCollector
}

type result struct {
	patches             map[utils.ValuesPatchType]pkg.OutputPatchableValuesCollector
	objectPathCollector pkg.OutputPatchCollector
	metricsCollector    pkg.OutputMetricsCollector
}

func (r *result) ValuesPatches() map[utils.ValuesPatchType]pkg.OutputPatchableValuesCollector {
	return r.patches
}

func (r *result) ObjectPatchCollector() pkg.OutputPatchCollector {
	return r.objectPathCollector
}

func (r *result) MetricsCollector() pkg.OutputMetricsCollector {
	return r.metricsCollector
}
