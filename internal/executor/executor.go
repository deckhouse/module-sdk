package executor

import (
	"context"

	bctx "github.com/deckhouse/module-sdk/internal/binding-context"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/utils"
)

type Executor interface {
	Config() *pkg.HookConfig
	Execute(ctx context.Context, req Request) (Result, error)
	IsApplicationHook() bool
}

type Request interface {
	GetValues() (map[string]any, error)
	GetConfigValues() (map[string]any, error)
	GetBindingContexts() ([]bctx.BindingContext, error)
	GetDependencyContainer() pkg.DependencyContainer
}

type Result interface {
	MetricsCollector() pkg.Outputer
	ObjectPatchCollector() pkg.Outputer
	ValuesPatchCollector(key utils.ValuesPatchType) pkg.Outputer
}

type result struct {
	objectPatchCollector pkg.Outputer
	metricsCollector     pkg.Outputer
	patches              map[utils.ValuesPatchType]pkg.Outputer
}

func (r *result) MetricsCollector() pkg.Outputer {
	return r.metricsCollector
}

func (r *result) ObjectPatchCollector() pkg.Outputer {
	return r.objectPatchCollector
}

func (r *result) ValuesPatchCollector(key utils.ValuesPatchType) pkg.Outputer {
	return r.patches[key]
}
