package hook

import (
	"context"
	"log/slog"

	"github.com/deckhouse/deckhouse/pkg/log"
	bindingcontext "github.com/deckhouse/module-sdk/internal/binding-context"
	metric "github.com/deckhouse/module-sdk/internal/metric"
	objectpatch "github.com/deckhouse/module-sdk/internal/object-patch"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/utils"
)

type GoHook struct {
	config        *pkg.HookConfig
	reconcileFunc pkg.ReconcileFunc

	logger *log.Logger
}

// NewGoHook creates a new go hook
func NewGoHook(config *pkg.HookConfig, f pkg.ReconcileFunc) *GoHook {
	logger := log.NewLogger(log.Options{})

	return &GoHook{
		config:        config,
		reconcileFunc: f,
		logger:        logger.Named("hook-auto-logger"),
	}
}

func (h *GoHook) GetName() string {
	return h.config.Metadata.Name
}

func (h *GoHook) GetPath() string {
	return h.config.Metadata.Path
}

func (h *GoHook) GetConfig() *pkg.HookConfig {
	return h.config
}

func (h *GoHook) SetMetadata(m *pkg.GoHookMetadata) {
	h.config.Metadata = *m
}

func (h *GoHook) SetLogger(logger *log.Logger) {
	h.logger = logger
}

type HookRequest interface {
	GetValues() (map[string]any, error)
	GetConfigValues() (map[string]any, error)
	GetBindingContexts() ([]bindingcontext.BindingContext, error)
	GetDependencyContainer() pkg.DependencyContainer
}

func (h *GoHook) Execute(ctx context.Context, req HookRequest) (*HookResult, error) {
	// Values are patched in-place, so an error can occur.
	rawValues, err := req.GetValues()
	if err != nil {
		h.logger.Error("get values", slog.String("error", err.Error()))
		return nil, err
	}

	patchableValues, err := NewPatchableValues(rawValues)
	if err != nil {
		h.logger.Error("new patchable values", slog.String("error", err.Error()))
		return nil, err
	}

	rawConfigValues, err := req.GetConfigValues()
	if err != nil {
		h.logger.Error("get config values", slog.String("error", err.Error()))
		return nil, err
	}

	patchableConfigValues, err := NewPatchableValues(rawConfigValues)
	if err != nil {
		h.logger.Error("new patchable config values", slog.String("error", err.Error()))
		return nil, err
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
	objectPatchCollector := objectpatch.NewObjectPatchCollector()

	err = h.Run(ctx, &pkg.HookInput{
		Snapshots:        formattedSnapshots,
		Values:           patchableValues,
		ConfigValues:     patchableConfigValues,
		PatchCollector:   objectPatchCollector,
		MetricsCollector: metricsCollector,
		DC:               req.GetDependencyContainer(),
		Logger:           h.logger.With("output", "gohook"),
	})
	if err != nil {
		return nil, err
	}

	return &HookResult{
		Patches: map[utils.ValuesPatchType]pkg.PatchableValuesCollector{
			utils.MemoryValuesPatch: patchableValues,
			utils.ConfigMapPatch:    patchableConfigValues,
		},
		Metrics:                 metricsCollector,
		ObjectPatcherOperations: objectPatchCollector,
	}, nil
}

// Run start ReconcileFunc
func (h *GoHook) Run(ctx context.Context, input *pkg.HookInput) error {
	return h.reconcileFunc(ctx, input)
}

// Bool returns a pointer to a bool.
func Bool(b bool) *bool {
	return &b
}

// BoolDeref dereferences the bool ptr and returns it if not nil, or else
// returns def.
func BoolDeref(ptr *bool, def bool) bool {
	if ptr != nil {
		return *ptr
	}
	return def
}

// HookResult returns result of a hook execution
type HookResult struct {
	Patches map[utils.ValuesPatchType]pkg.PatchableValuesCollector

	ObjectPatcherOperations pkg.PatchCollector
	Metrics                 pkg.MetricsCollector
}
