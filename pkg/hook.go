package pkg

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	EnvApplicationName      = "APPLICATION_NAME"
	EnvApplicationNamespace = "APPLICATION_NAMESPACE"
)

type Input interface {
	*HookInput | *ApplicationHookInput
}

type Hook[T Input] struct {
	Config   *HookConfig
	HookFunc HookFunc[T]
}

// HookFunc function which holds the main logic of the hook
type HookFunc[T Input] func(ctx context.Context, input T) error

type HookInput struct {
	Snapshots Snapshots

	Values           PatchableValuesCollector
	ConfigValues     PatchableValuesCollector
	PatchCollector   PatchCollector
	MetricsCollector MetricsCollector

	DC DependencyContainer

	Logger Logger
}

type ApplicationHookInput struct {
	Snapshots Snapshots

	Instance Instance

	Values           PatchableValuesCollector
	PatchCollector   NamespacedPatchCollector
	MetricsCollector MetricsCollector

	DC ApplicationDependencyContainer

	Logger Logger
}

// Instance in application instance getter
type Instance interface {
	// Name returns application instance name
	Name() string
	// Namespace returns application instance namespace
	Namespace() string
}

type HookMetadata struct {
	// Hook name
	Name string
	// Hook path
	Path string
}

// HookType defines the type of hook
type HookType string

const (
	HookTypeModule      HookType = "module"
	HookTypeApplication HookType = "application"
)

type HookConfig struct {
	Metadata              HookMetadata
	Schedule              []ScheduleConfig
	Kubernetes            []KubernetesConfig
	ApplicationKubernetes []ApplicationKubernetesConfig
	// OnStartup runs hook on module/global startup
	// Attention! During the startup you don't have snapshots available
	// use native KubeClient to fetch resources
	OnStartup         *OrderedConfig
	OnBeforeHelm      *OrderedConfig
	OnAfterHelm       *OrderedConfig
	OnAfterDeleteHelm *OrderedConfig

	AllowFailure bool
	Queue        string

	Settings *HookConfigSettings

	HookType HookType
}

var (
	camelCaseRegexp   = regexp.MustCompile(`^[a-zA-Z]*$`)
	cronScheduleRegex = regexp.MustCompile(`^((((\d+,)+\d+|(\d+(\/|-|#)\d+)|\d+L?|\*(\/\d+)?|L(-\d+)?|\?|[A-Z]{3}(-[A-Z]{3})?) ?){5,7})|(@(annually|yearly|monthly|weekly|daily|hourly|reboot))|(@every (\d+(ns|us|µs|ms|s|m|h))+)$`)
)

func (cfg *HookConfig) Validate() error {
	var errs error
	// list of not validated fields:
	// Metadata (filled by registry)
	for _, s := range cfg.Schedule {
		if err := s.Validate(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("schedule with name '%s': %w", s.Name, err))
		}
	}

	isApplicationHook := cfg.HookType == HookTypeApplication

	if isApplicationHook {
		if len(cfg.Kubernetes) > 0 {
			errs = errors.Join(errs, errors.New("application hooks must use ApplicationKubernetes field, not Kubernetes"))
		}
		for _, k := range cfg.ApplicationKubernetes {
			if err := k.Validate(); err != nil {
				errs = errors.Join(errs, fmt.Errorf("application kubernetes config with name '%s': %w", k.Name, err))
			}
		}
	} else {
		if len(cfg.ApplicationKubernetes) > 0 {
			errs = errors.Join(errs, errors.New("module hooks must use Kubernetes field, not ApplicationKubernetes"))
		}
		for _, k := range cfg.Kubernetes {
			if err := k.Validate(); err != nil {
				errs = errors.Join(errs, fmt.Errorf("kubernetes config with name '%s': %w", k.Name, err))
			}
		}
	}

	return errs
}

type OrderedConfig struct {
	Order uint
}

type HookConfigSettings struct {
	ExecutionMinInterval time.Duration
	ExecutionBurst       int
}

type ScheduleConfig struct {
	Name string
	// Crontab is a schedule config in crontab format. (5 or 6 fields)
	Crontab string
}

func (cfg *ScheduleConfig) Validate() error {
	var errs error

	if !cronScheduleRegex.Match([]byte(cfg.Crontab)) {
		errs = errors.Join(errs, errors.New("crontab is not valid"))
	}

	return errs
}

// KubernetesConfigInterface is a common interface for KubernetesConfig and ApplicationKubernetesConfig.
type KubernetesConfigInterface interface {
	GetName() string
	GetAPIVersion() string
	GetKind() string
	GetNameSelector() *NameSelector
	GetLabelSelector() *metav1.LabelSelector
	GetFieldSelector() *FieldSelector
	GetExecuteHookOnEvents() *bool
	GetExecuteHookOnSynchronization() *bool
	GetWaitForSynchronization() *bool
	GetJqFilter() string
	GetAllowFailure() *bool
	GetResynchronizationPeriod() string
	GetNamespaceSelector() *NamespaceSelector // Returns nil for ApplicationKubernetesConfig
}

// ApplicationKubernetesConfig is used for application hooks.
// Application hooks automatically work in the application's namespace,
// so NamespaceSelector is not allowed.
type ApplicationKubernetesConfig struct {
	// Name is a key in snapshots map.
	Name string
	// APIVersion of objects. "v1" is used if not set.
	APIVersion string
	// Kind of objects.
	Kind string
	// NameSelector used to subscribe on object by its name.
	NameSelector *NameSelector
	// LabelSelector used to subscribe on objects by matching their labels.
	LabelSelector *metav1.LabelSelector
	// FieldSelector used to subscribe on objects by matching specific fields (the list of fields is narrow, see shell-operator documentation).
	FieldSelector *FieldSelector
	// ExecuteHookOnEvents is true by default. Set to false if only snapshot update is needed.
	ExecuteHookOnEvents *bool
	// ExecuteHookOnSynchronization is true by default. Set to false if only snapshot update is needed.
	ExecuteHookOnSynchronization *bool
	// WaitForSynchronization is true by default. Set to false if beforeHelm is not required this snapshot on start.
	WaitForSynchronization *bool
	// JQ filter to filter results from kubernetes objects
	JqFilter string
	// Allow to fail hook
	AllowFailure *bool

	ResynchronizationPeriod string
}

type KubernetesConfig struct {
	// Name is a key in snapshots map.
	Name string
	// APIVersion of objects. "v1" is used if not set.
	APIVersion string
	// Kind of objects.
	Kind string
	// NameSelector used to subscribe on object by its name.
	NameSelector *NameSelector
	// NamespaceSelector used to subscribe on objects in namespaces.
	NamespaceSelector *NamespaceSelector
	// LabelSelector used to subscribe on objects by matching their labels.
	LabelSelector *metav1.LabelSelector
	// FieldSelector used to subscribe on objects by matching specific fields (the list of fields is narrow, see shell-operator documentation).
	FieldSelector *FieldSelector
	// ExecuteHookOnEvents is true by default. Set to false if only snapshot update is needed.
	ExecuteHookOnEvents *bool
	// ExecuteHookOnSynchronization is true by default. Set to false if only snapshot update is needed.
	ExecuteHookOnSynchronization *bool
	// WaitForSynchronization is true by default. Set to false if beforeHelm is not required this snapshot on start.
	WaitForSynchronization *bool
	// JQ filter to filter results from kubernetes objects
	JqFilter string
	// Allow to fail hook
	AllowFailure *bool

	ResynchronizationPeriod string
}

// Implement KubernetesConfigInterface for ApplicationKubernetesConfig
func (cfg *ApplicationKubernetesConfig) GetName() string {
	return cfg.Name
}

func (cfg *ApplicationKubernetesConfig) GetAPIVersion() string {
	return cfg.APIVersion
}

func (cfg *ApplicationKubernetesConfig) GetKind() string {
	return cfg.Kind
}

func (cfg *ApplicationKubernetesConfig) GetNameSelector() *NameSelector {
	return cfg.NameSelector
}

func (cfg *ApplicationKubernetesConfig) GetLabelSelector() *metav1.LabelSelector {
	return cfg.LabelSelector
}

func (cfg *ApplicationKubernetesConfig) GetFieldSelector() *FieldSelector {
	return cfg.FieldSelector
}

func (cfg *ApplicationKubernetesConfig) GetExecuteHookOnEvents() *bool {
	return cfg.ExecuteHookOnEvents
}

func (cfg *ApplicationKubernetesConfig) GetExecuteHookOnSynchronization() *bool {
	return cfg.ExecuteHookOnSynchronization
}

func (cfg *ApplicationKubernetesConfig) GetWaitForSynchronization() *bool {
	return cfg.WaitForSynchronization
}

func (cfg *ApplicationKubernetesConfig) GetJqFilter() string {
	return cfg.JqFilter
}

func (cfg *ApplicationKubernetesConfig) GetAllowFailure() *bool {
	return cfg.AllowFailure
}

func (cfg *ApplicationKubernetesConfig) GetResynchronizationPeriod() string {
	return cfg.ResynchronizationPeriod
}

func (cfg *ApplicationKubernetesConfig) GetNamespaceSelector() *NamespaceSelector {
	return nil // Application hooks don't have namespace selector
}

// you must test JqFilter by yourself
func (cfg *ApplicationKubernetesConfig) Validate() error {
	var errs error

	if !camelCaseRegexp.Match([]byte(cfg.Kind)) {
		errs = errors.Join(errs, errors.New("kind has not letter symbols"))
	}

	return errs
}

// you must test JqFilter by yourself
func (cfg *KubernetesConfig) Validate() error {
	var errs error

	if !camelCaseRegexp.Match([]byte(cfg.Kind)) {
		errs = errors.Join(errs, errors.New("kind has not letter symbols"))
	}

	return errs
}

// Implement KubernetesConfigInterface for KubernetesConfig
func (cfg *KubernetesConfig) GetName() string {
	return cfg.Name
}

func (cfg *KubernetesConfig) GetAPIVersion() string {
	return cfg.APIVersion
}

func (cfg *KubernetesConfig) GetKind() string {
	return cfg.Kind
}

func (cfg *KubernetesConfig) GetNameSelector() *NameSelector {
	return cfg.NameSelector
}

func (cfg *KubernetesConfig) GetLabelSelector() *metav1.LabelSelector {
	return cfg.LabelSelector
}

func (cfg *KubernetesConfig) GetFieldSelector() *FieldSelector {
	return cfg.FieldSelector
}

func (cfg *KubernetesConfig) GetExecuteHookOnEvents() *bool {
	return cfg.ExecuteHookOnEvents
}

func (cfg *KubernetesConfig) GetExecuteHookOnSynchronization() *bool {
	return cfg.ExecuteHookOnSynchronization
}

func (cfg *KubernetesConfig) GetWaitForSynchronization() *bool {
	return cfg.WaitForSynchronization
}

func (cfg *KubernetesConfig) GetJqFilter() string {
	return cfg.JqFilter
}

func (cfg *KubernetesConfig) GetAllowFailure() *bool {
	return cfg.AllowFailure
}

func (cfg *KubernetesConfig) GetResynchronizationPeriod() string {
	return cfg.ResynchronizationPeriod
}

func (cfg *KubernetesConfig) GetNamespaceSelector() *NamespaceSelector {
	return cfg.NamespaceSelector
}

type NameSelector struct {
	MatchNames []string
}

type FieldSelectorRequirement struct {
	Field    string
	Operator string
	Value    string
}

type FieldSelector struct {
	MatchExpressions []FieldSelectorRequirement
}

type NamespaceSelector struct {
	NameSelector  *NameSelector
	LabelSelector *metav1.LabelSelector
}
