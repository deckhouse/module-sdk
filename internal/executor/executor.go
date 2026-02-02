package executor

import (
	"context"

	bctx "github.com/deckhouse/module-sdk/internal/binding-context"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/utils"
)

// Executor runs a hook with the provided request and returns results.
// Implemented by moduleExecutor and applicationExecutor.
type Executor interface {
	// Config returns the hook's configuration (*HookConfig or *ApplicationHookConfig).
	Config() any
	// Execute runs the hook logic and returns collected results.
	Execute(ctx context.Context, req Request) (Result, error)
}

// Request provides input data for hook execution.
// Implemented by transport layer (e.g., file transport).
type Request interface {
	// GetValues returns the current module values.
	GetValues() (map[string]any, error)
	// GetConfigValues returns the module's ConfigMap values.
	GetConfigValues() (map[string]any, error)
	// GetBindingContexts returns Kubernetes binding contexts with snapshots.
	GetBindingContexts() ([]bctx.BindingContext, error)
	// GetDependencyContainer returns the container with external dependencies.
	GetDependencyContainer() pkg.DependencyContainer
}

// Result contains outputs collected during hook execution.
// Used by transport layer to send results back to shell-operator.
type Result interface {
	// MetricsCollector returns collected Prometheus metrics.
	MetricsCollector() pkg.Outputer
	// ObjectPatchCollector returns collected Kubernetes object patches.
	ObjectPatchCollector() pkg.Outputer
	// ValuesPatchCollector returns collected values patches by type.
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
