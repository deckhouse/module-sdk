package executor

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/internal/metric"
	"github.com/deckhouse/module-sdk/internal/objectpatch"
	"github.com/deckhouse/module-sdk/pkg"
	patchablevalues "github.com/deckhouse/module-sdk/pkg/patchable-values"
	"github.com/deckhouse/module-sdk/pkg/utils"
)

type applicationExecutor struct {
	hook   pkg.Hook[*pkg.ApplicationHookInput]
	logger *log.Logger
}

// NewApplicationExecutor creates a new application executor
func NewApplicationExecutor(h pkg.Hook[*pkg.ApplicationHookInput], logger *log.Logger) Executor {
	return &applicationExecutor{
		hook:   h,
		logger: logger,
	}
}

func (e *applicationExecutor) Config() *pkg.HookConfig {
	return e.hook.Config
}

func (e *applicationExecutor) Execute(ctx context.Context, req Request) (Result, error) {
	// Values are patched in-place, so an error can occur.
	rawValues, err := req.GetValues()
	if err != nil {
		e.logger.Error("get values", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get values: %w", err)
	}

	patchableValues, err := patchablevalues.NewPatchableValues(rawValues)
	if err != nil {
		e.logger.Error("new patchable values", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get patchable values: %w", err)
	}

	bContext, err := req.GetBindingContexts()
	if err != nil {
		e.logger.Warn("get binding context", slog.String("error", err.Error()))
	}

	formattedSnapshots := make(objectpatch.Snapshots, len(bContext))
	for _, bc := range bContext {
		for snapBindingName, snaps := range bc.Snapshots {
			for _, snap := range snaps {
				if snap.FilterResult != nil {
					formattedSnapshots[snapBindingName] = append(formattedSnapshots[snapBindingName], objectpatch.Snapshot(snap.FilterResult))

					continue
				}

				if snap.Object != nil {
					formattedSnapshots[snapBindingName] = append(formattedSnapshots[snapBindingName], objectpatch.Snapshot(snap.Object))

					continue
				}
			}
		}
	}

	inst := newAppInstance()

	metricsCollector := metric.NewCollector()
	namespacedPatchCollector := objectpatch.NewNamespacedCollector(inst.namespace, e.logger.Named("object-patch-collector"))

	dc, ok := req.GetDependencyContainer().(pkg.ApplicationDependencyContainer)
	if !ok {
		e.logger.Error("get application dependency container", slog.String("error", "request dependency container is not an ApplicationDependencyContainer"))
		return nil, fmt.Errorf("get application dependency container: incompatible dependency container type")
	}

	err = e.hook.HookFunc(ctx, &pkg.ApplicationHookInput{
		Snapshots:        formattedSnapshots,
		Instance:         inst,
		Values:           patchableValues,
		PatchCollector:   namespacedPatchCollector,
		MetricsCollector: metricsCollector,
		DC:               dc,
		Logger:           e.logger,
	})
	if err != nil {
		return nil, fmt.Errorf("hook reconcile func: %w", err)
	}

	return &result{
		patches: map[utils.ValuesPatchType]pkg.Outputer{
			utils.MemoryValuesPatch: patchableValues,
		},
		objectPatchCollector: namespacedPatchCollector,
		metricsCollector:     metricsCollector,
	}, nil
}

type applicationInstance struct {
	name      string
	namespace string
}

func newAppInstance() *applicationInstance {
	return &applicationInstance{
		name:      os.Getenv(pkg.EnvApplicationName),
		namespace: os.Getenv(pkg.EnvApplicationNamespace),
	}
}

func (i *applicationInstance) Name() string {
	return i.name
}

func (i *applicationInstance) Namespace() string {
	return i.namespace
}
