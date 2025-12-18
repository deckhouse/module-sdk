package executor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/internal/metric"
	objectpatch "github.com/deckhouse/module-sdk/internal/object-patch"
	"github.com/deckhouse/module-sdk/pkg"
	patchablevalues "github.com/deckhouse/module-sdk/pkg/patchable-values"
	"github.com/deckhouse/module-sdk/pkg/utils"
)

type applicationExecutor struct {
	hook   pkg.Hook[*pkg.ApplicationHookInput]
	logger *log.Logger
}

// NewApplicationExecutor creates a new application hook
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

	metricsCollector := metric.NewCollector()
	objectPatchCollector := objectpatch.NewObjectPatchCollector(e.logger.Named("object-patch-collector"))

	err = e.hook.HookFunc(ctx, &pkg.ApplicationHookInput{
		Snapshots:        formattedSnapshots,
		Values:           patchableValues,
		PatchCollector:   objectPatchCollector,
		MetricsCollector: metricsCollector,
		DC:               req.GetDependencyContainer(),
		Logger:           e.logger,
	})
	if err != nil {
		return nil, fmt.Errorf("hook reconcile func: %w", err)
	}

	return &result{
		patches: map[utils.ValuesPatchType]pkg.OutputPatchableValuesCollector{
			utils.MemoryValuesPatch: patchableValues,
		},
		objectPathCollector: objectPatchCollector,
		metricsCollector:    metricsCollector,
	}, nil
}
