package pkg

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Hook struct {
	Config        *HookConfig
	ReconcileFunc ReconcileFunc
}

// ReconcileFunc function which holds the main logic of the hook
type ReconcileFunc func(ctx context.Context, input *HookInput) error

type HookInput struct {
	Snapshots Snapshots

	Values           PatchableValuesCollector
	ConfigValues     PatchableValuesCollector
	PatchCollector   PatchCollector
	MetricsCollector MetricsCollector

	DC DependencyContainer

	Logger Logger
}

type HookMetadata struct {
	// Hook name
	Name string
	// Hook path
	Path string
}

type HookConfig struct {
	Metadata   HookMetadata
	Schedule   []ScheduleConfig
	Kubernetes []KubernetesConfig
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
}

var (
	kebabCaseRegexp   = regexp.MustCompile(`^[a-z]-$`)
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

	for _, k := range cfg.Kubernetes {
		if err := k.Validate(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("kubernetes config with name '%s': %w", k.Name, err))
		}
	}

	return nil
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

	if !camelCaseRegexp.Match([]byte(cfg.Name)) {
		errs = errors.Join(errs, errors.New("name has not letter symbols"))
	}

	if !cronScheduleRegex.Match([]byte(cfg.Crontab)) {
		errs = errors.Join(errs, errors.New("crontab is not valid"))
	}

	return errs
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

// you must test JqFilter by yourself
func (cfg *KubernetesConfig) Validate() error {
	var errs error

	if !kebabCaseRegexp.Match([]byte(cfg.Name)) {
		errs = errors.Join(errs, errors.New("name is not kebab case"))
	}

	if !camelCaseRegexp.Match([]byte(cfg.Kind)) {
		errs = errors.Join(errs, errors.New("kind has not letter symbols"))
	}

	if err := cfg.NameSelector.Validate(); err != nil {
		errs = errors.Join(errs, fmt.Errorf("name selector: %w", err))
	}

	if err := cfg.NamespaceSelector.Validate(); err != nil {
		errs = errors.Join(errs, fmt.Errorf("namespace selector: %w", err))
	}

	return errs
}

type NameSelector struct {
	MatchNames []string
}

func (cfg *NameSelector) Validate() error {
	if cfg == nil {
		return nil
	}

	var errs error

	for _, sel := range cfg.MatchNames {
		if !kebabCaseRegexp.Match([]byte(sel)) {
			errs = errors.Join(errs, fmt.Errorf("selector is not kebab case '%s'", sel))
		}
	}

	return errs
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

func (cfg *NamespaceSelector) Validate() error {
	if cfg == nil {
		return nil
	}

	var errs error

	if err := cfg.NameSelector.Validate(); err != nil {
		errs = errors.Join(errs, fmt.Errorf("name selector: %w", err))
	}

	return errs
}
