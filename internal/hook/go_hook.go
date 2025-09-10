package hook

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/deckhouse/deckhouse/pkg/log"

	bindingcontext "github.com/deckhouse/module-sdk/internal/binding-context"
	metric "github.com/deckhouse/module-sdk/internal/metric"
	objectpatch "github.com/deckhouse/module-sdk/internal/object-patch"
	"github.com/deckhouse/module-sdk/pkg"
	patchablevalues "github.com/deckhouse/module-sdk/pkg/patchable-values"
	"github.com/deckhouse/module-sdk/pkg/utils"
)

type Hook struct {
	config        *pkg.HookConfig
	reconcileFunc pkg.ReconcileFunc

	logger *log.Logger
}

// NewHook creates a new go hook
func NewHook(config *pkg.HookConfig, f pkg.ReconcileFunc) *Hook {
	logger := log.NewLogger()

	return &Hook{
		config:        config,
		reconcileFunc: f,
		logger:        logger.Named("hook-auto-logger"),
	}
}

func (h *Hook) GetName() string {
	return h.config.Metadata.Name
}

func (h *Hook) GetPath() string {
	return h.config.Metadata.Path
}

func (h *Hook) GetConfig() *pkg.HookConfig {
	return h.config
}

func (h *Hook) SetMetadata(m *pkg.HookMetadata) *Hook {
	h.config.Metadata = *m

	return h
}

func (h *Hook) SetLogger(logger *log.Logger) *Hook {
	h.logger = logger

	return h
}

type HookRequest interface {
	GetValues() (map[string]any, error)
	GetConfigValues() (map[string]any, error)
	GetBindingContexts() ([]bindingcontext.BindingContext, error)
	GetDependencyContainer() pkg.DependencyContainer
}

func (h *Hook) Execute(ctx context.Context, req HookRequest) (*HookResult, error) {
	// Values are patched in-place, so an error can occur.
	rawValues, err := req.GetValues()
	if err != nil {
		h.logger.Error("get values", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get values: %w", err)
	}

	patchableValues, err := patchablevalues.NewPatchableValues(rawValues)
	if err != nil {
		h.logger.Error("new patchable values", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get patchable values: %w", err)
	}

	rawConfigValues, err := req.GetConfigValues()
	if err != nil {
		h.logger.Error("get config values", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get config values: %w", err)
	}

	patchableConfigValues, err := patchablevalues.NewPatchableValues(rawConfigValues)
	if err != nil {
		h.logger.Error("new patchable config values", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get patchable config values: %w", err)
	}

	bContext, err := req.GetBindingContexts()
	if err != nil {
		h.logger.Warn("get binding context", slog.String("error", err.Error()))
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
	objectPatchCollector := objectpatch.NewObjectPatchCollector(h.logger.Named("object-patch-collector"))

	err = h.Run(ctx, &pkg.HookInput{
		Snapshots:        formattedSnapshots,
		Values:           patchableValues,
		ConfigValues:     patchableConfigValues,
		PatchCollector:   objectPatchCollector,
		MetricsCollector: metricsCollector,
		DC:               req.GetDependencyContainer(),
		Logger:           h.logger,
	})
	if err != nil {
		return nil, fmt.Errorf("hook reconcile func: %w", err)
	}

	return &HookResult{
		Patches: map[utils.ValuesPatchType]pkg.OutputPatchableValuesCollector{
			utils.MemoryValuesPatch: patchableValues,
			utils.ConfigMapPatch:    patchableConfigValues,
		},
		Metrics:                 metricsCollector,
		ObjectPatcherOperations: objectPatchCollector,
	}, nil
}

// Run start ReconcileFunc
func (h *Hook) Run(ctx context.Context, input *pkg.HookInput) error {
	return h.reconcileFunc(ctx, input)
}

// HookResult returns result of a hook execution
type HookResult struct {
	Patches map[utils.ValuesPatchType]pkg.OutputPatchableValuesCollector

	ObjectPatcherOperations pkg.OutputPatchCollector
	Metrics                 pkg.OutputMetricsCollector
}
