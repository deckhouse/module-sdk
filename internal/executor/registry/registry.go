package registry

import (
	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/internal/executor"
	"github.com/deckhouse/module-sdk/pkg"
)

type Registry struct {
	executors         []executor.Executor
	readinessExecutor executor.Executor

	logger *log.Logger
}

func NewRegistry(logger *log.Logger) *Registry {
	return &Registry{
		executors: make([]executor.Executor, 0, 1),
		logger:    logger,
	}
}

// Executors returns all executors
func (r *Registry) Executors() []executor.Executor {
	return r.executors
}

// Readiness returns the readiness hook
// It is used to check if the module is ready to serve requests
// It is not used for the readiness probe
// The readiness probe is implemented in the module itself
func (r *Registry) Readiness() executor.Executor {
	return r.readinessExecutor
}

func (r *Registry) RegisterModuleHooks(hooks ...pkg.Hook[*pkg.HookInput]) {
	for _, h := range hooks {
		exec := executor.NewModuleExecutor(h, r.logger.Named(h.Config.Metadata.Name))
		r.executors = append(r.executors, exec)
	}
}

func (r *Registry) RegisterAppHooks(hooks ...pkg.Hook[*pkg.ApplicationHookInput]) {
	for _, h := range hooks {
		exec := executor.NewApplicationExecutor(h, r.logger.Named(h.Config.Metadata.Name))
		r.executors = append(r.executors, exec)
	}
}

func (r *Registry) SetReadinessHook(h pkg.Hook[*pkg.HookInput]) {
	r.readinessExecutor = executor.NewModuleExecutor(h, r.logger.Named(h.Config.Metadata.Name))
}
