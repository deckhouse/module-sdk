package pkg

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// =============================================================================
// Constants and Variables
// =============================================================================

const (
	EnvApplicationName      = "APPLICATION_NAME"
	EnvApplicationNamespace = "APPLICATION_NAMESPACE"
)

var (
	camelCaseRegexp   = regexp.MustCompile(`^[a-zA-Z]*$`)
	cronScheduleRegex = regexp.MustCompile(`^((((\d+,)+\d+|(\d+(\/|-|#)\d+)|\d+L?|\*(\/\d+)?|L(-\d+)?|\?|[A-Z]{3}(-[A-Z]{3})?) ?){5,7})|(@(annually|yearly|monthly|weekly|daily|hourly|reboot))|(@every (\d+(ns|us|µs|ms|s|m|h))+)$`)
)

// =============================================================================
// Core Types and Type Constraints
// =============================================================================

// Input is a type constraint for hook input types.
type Input interface {
	*HookInput | *ApplicationHookInput
}

// Config is a type constraint for hook configuration types (used in generics).
type Config interface {
	HookConfig | ApplicationHookConfig | *HookConfig | *ApplicationHookConfig
}

// Hook represents a generic hook with configuration and execution function.
type Hook[C Config, T Input] struct {
	Config   C
	HookFunc HookFunc[T]
}

// HookFunc is the function signature for hook execution logic.
type HookFunc[T Input] func(ctx context.Context, input T) error

// =============================================================================
// Hook Input Types
// =============================================================================

// HookInput provides context and utilities for module hook execution.
type HookInput struct {
	Snapshots Snapshots

	Values           PatchableValuesCollector
	ConfigValues     PatchableValuesCollector
	PatchCollector   PatchCollector
	MetricsCollector MetricsCollector

	DC DependencyContainer

	Logger Logger
}

// ApplicationHookInput provides context and utilities for application hook execution.
type ApplicationHookInput struct {
	Snapshots Snapshots

	Instance Instance

	Values           PatchableValuesCollector
	PatchCollector   NamespacedPatchCollector
	MetricsCollector MetricsCollector

	DC ApplicationDependencyContainer

	Logger Logger
}

// Instance provides access to application instance metadata.
type Instance interface {
	// Name returns application instance name
	Name() string
	// Namespace returns application instance namespace
	Namespace() string
}

// =============================================================================
// Hook Metadata and Settings
// =============================================================================

// HookMetadata contains identifying information for a hook.
type HookMetadata struct {
	// Name is the hook's unique identifier
	Name string
	// Path is the file path where the hook is defined
	Path string
}

// HookConfigInterface is implemented by *HookConfig and *ApplicationHookConfig.
// It provides type-safe access to common config fields and explicit conversion
// to concrete types, avoiding unsafe type assertions on any.
type HookConfigInterface interface {
	GetMetadata() HookMetadata
	GetQueue() string
	// AsHookConfig returns the config as *HookConfig if it is a module hook config.
	AsHookConfig() (*HookConfig, bool)
	// AsApplicationHookConfig returns the config as *ApplicationHookConfig if it is an application hook config.
	AsApplicationHookConfig() (*ApplicationHookConfig, bool)
}

// OrderedConfig specifies execution order for lifecycle hooks.
type OrderedConfig struct {
	Order uint
}

// HookConfigSettings contains rate limiting settings for hook execution.
type HookConfigSettings struct {
	ExecutionMinInterval time.Duration
	ExecutionBurst       int
}

// =============================================================================
// Module Hook Configuration
// =============================================================================

// HookConfig defines the configuration for a module hook.
type HookConfig struct {
	Metadata HookMetadata
	Schedule []ScheduleConfig

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

// Validate checks the HookConfig for errors.
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

	return errs
}

// GetMetadata implements HookConfigLike.
func (cfg *HookConfig) GetMetadata() HookMetadata { return cfg.Metadata }

// GetQueue implements HookConfigLike.
func (cfg *HookConfig) GetQueue() string { return cfg.Queue }

// AsHookConfig implements HookConfigLike.
func (cfg *HookConfig) AsHookConfig() (*HookConfig, bool) { return cfg, true }

// AsApplicationHookConfig implements HookConfigLike.
func (cfg *HookConfig) AsApplicationHookConfig() (*ApplicationHookConfig, bool) { return nil, false }

// =============================================================================
// Application Hook Configuration
// =============================================================================

// ApplicationHookConfig defines the configuration for an application hook.
type ApplicationHookConfig struct {
	Metadata HookMetadata
	Schedule []ScheduleConfig

	Kubernetes []ApplicationKubernetesConfig

	// OnStartup runs hook on application startup
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

// Validate checks the ApplicationHookConfig for errors.
func (cfg *ApplicationHookConfig) Validate() error {
	var errs error
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

	return errs
}

// GetMetadata implements HookConfigLike.
func (cfg *ApplicationHookConfig) GetMetadata() HookMetadata { return cfg.Metadata }

// GetQueue implements HookConfigLike.
func (cfg *ApplicationHookConfig) GetQueue() string { return cfg.Queue }

// AsHookConfig implements HookConfigLike.
func (cfg *ApplicationHookConfig) AsHookConfig() (*HookConfig, bool) { return nil, false }

// AsApplicationHookConfig implements HookConfigLike.
func (cfg *ApplicationHookConfig) AsApplicationHookConfig() (*ApplicationHookConfig, bool) {
	return cfg, true
}

// =============================================================================
// Schedule Configuration
// =============================================================================

// ScheduleConfig defines a cron-based schedule for hook execution.
type ScheduleConfig struct {
	Name string
	// Crontab is a schedule config in crontab format. (5 or 6 fields)
	Crontab string
}

// Validate checks the ScheduleConfig for errors.
func (cfg *ScheduleConfig) Validate() error {
	var errs error

	if !cronScheduleRegex.Match([]byte(cfg.Crontab)) {
		errs = errors.Join(errs, errors.New("crontab is not valid"))
	}

	return errs
}

// =============================================================================
// Kubernetes Configuration
// =============================================================================

// KubernetesConfig defines a subscription to Kubernetes objects for module hooks.
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
	// JqFilter filters results from kubernetes objects.
	JqFilter string
	// AllowFailure allows the hook to fail without stopping execution.
	AllowFailure *bool

	ResynchronizationPeriod string
}

// Validate checks the KubernetesConfig for errors.
func (cfg *KubernetesConfig) Validate() error {
	var errs error

	if !camelCaseRegexp.Match([]byte(cfg.Kind)) {
		errs = errors.Join(errs, errors.New("kind has not letter symbols"))
	}

	return errs
}

// ApplicationKubernetesConfig defines a subscription to Kubernetes objects for application hooks.
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
	// JqFilter filters results from kubernetes objects.
	JqFilter string
	// AllowFailure allows the hook to fail without stopping execution.
	AllowFailure *bool

	ResynchronizationPeriod string
}

// Validate checks the ApplicationKubernetesConfig for errors.
func (cfg *ApplicationKubernetesConfig) Validate() error {
	var errs error

	if !camelCaseRegexp.Match([]byte(cfg.Kind)) {
		errs = errors.Join(errs, errors.New("kind has not letter symbols"))
	}

	return errs
}

// =============================================================================
// Selectors
// =============================================================================

// NameSelector filters objects by name.
type NameSelector struct {
	MatchNames []string
}

// FieldSelectorRequirement defines a single field selector condition.
type FieldSelectorRequirement struct {
	Field    string
	Operator string
	Value    string
}

// FieldSelector filters objects by field values.
type FieldSelector struct {
	MatchExpressions []FieldSelectorRequirement
}

// NamespaceSelector filters namespaces for object subscription.
type NamespaceSelector struct {
	NameSelector  *NameSelector
	LabelSelector *metav1.LabelSelector
}
