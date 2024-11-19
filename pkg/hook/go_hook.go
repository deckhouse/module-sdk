package hook

import (
	"log/slog"

	"github.com/deckhouse/deckhouse/pkg/log"
	service "github.com/deckhouse/module-sdk/pkg"
	bindingcontext "github.com/deckhouse/module-sdk/pkg/binding-context"
	"github.com/deckhouse/module-sdk/pkg/hook/config"
	"github.com/deckhouse/module-sdk/pkg/kubernetes"
	metric "github.com/deckhouse/module-sdk/pkg/metric"
	"github.com/deckhouse/module-sdk/pkg/utils"
)

type GoHook struct {
	// Metadata?
	Name string
	Path string

	config        *config.HookConfig
	reconcileFunc ReconcileFunc

	logger *log.Logger
}

// NewGoHook creates a new go hook
func NewGoHook(config *config.HookConfig, f ReconcileFunc) *GoHook {
	logger := log.NewLogger(log.Options{
		Level: log.LogLevelFromStr(config.LogLevelRaw).Level(),
	})

	return &GoHook{
		config:        config,
		reconcileFunc: f,
		logger:        logger.Named("hook-auto-logger"),
	}
}

func (h *GoHook) GetName() string {
	return h.Name
}

func (h *GoHook) GetPath() string {
	return h.Path
}

func (h *GoHook) GetConfig() *config.HookConfig {
	return h.config
}

func (h *GoHook) SetLogger(logger *log.Logger) {
	h.logger = logger
}

func (h *GoHook) Execute() (*HookResult, error) {
	// Values are patched in-place, so an error can occur.
	patchableValues, err := NewPatchableValues(utils.GetValues())
	if err != nil {
		h.logger.Error("new patchable values", slog.String("error", err.Error()))
		return nil, err
	}

	patchableConfigValues, err := NewPatchableValues(utils.GetConfigValues())
	if err != nil {
		h.logger.Error("new patchable config values", slog.String("error", err.Error()))
		return nil, err
	}

	bindingActions := make([]BindingAction, 0, 1)

	bContext, err := bindingcontext.GetBindingContexts()
	if err != nil {
		h.logger.Error("get binding context", slog.String("error", err.Error()))
	}

	formattedSnapshots := make(Snapshots, len(bContext))
	for _, bc := range bContext {
		for snapBindingName, snaps := range bc.Snapshots {
			formattedSnapshots[snapBindingName] = append(formattedSnapshots[snapBindingName], snaps...)
		}
	}

	metricsCollector := metric.NewCollector()
	objectPatchCollector := kubernetes.NewObjectPatchCollector()

	err = h.Run(&HookInput{
		Snapshots:       formattedSnapshots,
		Values:          patchableValues,
		ConfigValues:    patchableConfigValues,
		PatchCollector:  objectPatchCollector,
		MetricCollector: metricsCollector,
		BindingContexts: bindingActions,
		Logger:          h.logger.With("output", "gohook"),
	})
	if err != nil {
		return nil, err
	}

	return &HookResult{
		Patches: map[utils.ValuesPatchType]service.PatchableValuesCollector{
			utils.MemoryValuesPatch: patchableValues,
			utils.ConfigMapPatch:    patchableConfigValues,
		},
		Metrics:                 metricsCollector,
		ObjectPatcherOperations: objectPatchCollector,
		BindingActions:          bindingActions,
	}, nil
}

// Run start ReconcileFunc
func (h *GoHook) Run(input *HookInput) error {
	return h.reconcileFunc(input)
}

type Snapshots map[string][]kubernetes.ObjectAndFilterResult

type HookInput struct {
	Snapshots Snapshots

	Values          service.PatchableValuesCollector
	ConfigValues    service.PatchableValuesCollector
	PatchCollector  service.PatchCollector
	MetricCollector service.MetricCollector

	BindingContexts []BindingAction
	Logger          service.Logger
}

type BindingAction struct {
	Name       string // binding name
	Action     string // Disable / UpdateKind
	Kind       string
	ApiVersion string
}

// ReconcileFunc function which holds the main logic of the hook
type ReconcileFunc func(input *HookInput) error

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
	Patches map[utils.ValuesPatchType]service.PatchableValuesCollector

	ObjectPatcherOperations service.PatchCollector
	Metrics                 service.MetricCollector

	BindingActions []BindingAction
}
