package hook

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type GoHookMetadata struct {
	// Name is a key in snapshots map.
	Name string `yaml:"name" json:"name"`
	// Name is path to hook build folder.
	// Use it only if you need to build unique hook name
	Path string `yaml:"path" json:"path"`
}

const BatchHookConfigV1 = "v1"

type BatchHookConfig struct {
	Version          string       `yaml:"version" json:"version"`
	Hooks            []HookConfig `yaml:"hooks" json:"hooks"`
	Readiness        *HookConfig  `yaml:"readiness,omitempty" json:"readiness,omitempty"`
	HasSettingsCheck bool         `yaml:"has_settings_check,omitempty" json:"has_settings_check,omitempty"`
}

type HookConfig struct {
	ConfigVersion string             `yaml:"configVersion" json:"configVersion"`
	Metadata      GoHookMetadata     `yaml:"metadata" json:"metadata"`
	Schedule      []ScheduleConfig   `yaml:"schedule,omitempty" json:"schedule,omitempty"`
	Kubernetes    []KubernetesConfig `yaml:"kubernetes,omitempty" json:"kubernetes,omitempty"`
	// OnStartup runs hook on module/global startup
	// Attention! During the startup you don't have snapshots available
	// use native KubeClient to fetch resources
	OnStartup         *uint `yaml:"onStartup,omitempty" json:"onStartup,omitempty"`
	OnBeforeHelm      *uint `yaml:"beforeHelm,omitempty" json:"beforeHelm,omitempty"`
	OnAfterHelm       *uint `yaml:"afterHelm,omitempty" json:"afterHelm,omitempty"`
	OnAfterDeleteHelm *uint `yaml:"afterDeleteHelm,omitempty" json:"afterDeleteHelm,omitempty"`
	AllowFailure      *bool `yaml:"allowFailure,omitempty" json:"allowFailure,omitempty"`

	Settings *HookConfigSettings `yaml:"settings,omitempty" json:"settings,omitempty"`

	LogLevelRaw string `yaml:"logLevel,omitempty" json:"logLevel,omitempty"`
}

type ReadinessConfig struct {
	// ModuleName is a name of module to check
	ModuleName string `yaml:"moduleName" json:"moduleName"`
	// IntervalInSeconds is a time interval between checks
	// Default value is 1 second
	IntervalInSeconds int `yaml:"interval,omitempty" json:"interval,omitempty"`
}

type HookConfigSettings struct {
	ExecutionMinInterval time.Duration `yaml:"executionMinInterval,omitempty" json:"executionMinInterval,omitempty"`
	ExecutionBurst       int           `yaml:"executionBurst,omitempty" json:"executionBurst,omitempty"`
	// EnableSchedulesOnStartup
	// set to true, if you need to run 'Schedule' hooks without waiting addon-operator readiness
	EnableSchedulesOnStartup *bool `yaml:"enableSchedulesOnStartup,omitempty" json:"enableSchedulesOnStartup,omitempty"`
}

type ScheduleConfig struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Crontab is a schedule config in crontab format. (5 or 6 fields)
	Crontab string `yaml:"crontab" json:"crontab"`

	Queue string `yaml:"queue" json:"queue,omitempty"`
}

type FilterResult any

type FilterFunc func(*unstructured.Unstructured) (FilterResult, error)

type NameSelector struct {
	MatchNames []string `yaml:"matchNames" json:"matchNames"`
}

type FieldSelectorRequirement struct {
	Field    string `yaml:"field" json:"field"`
	Operator string `yaml:"operator" json:"operator"`
	Value    string `yaml:"value,omitempty" json:"value,omitempty"`
}

type FieldSelector struct {
	MatchExpressions []FieldSelectorRequirement `yaml:"matchExpressions" json:"matchExpressions"`
}

type NamespaceSelector struct {
	NameSelector  *NameSelector         `yaml:"nameSelector,omitempty" json:"nameSelector,omitempty"`
	LabelSelector *metav1.LabelSelector `yaml:"labelSelector,omitempty" json:"labelSelector,omitempty"`
}

type KubernetesConfig struct {
	Name string `yaml:"name" json:"name"`
	// APIVersion of objects. "v1" is used if not set.
	APIVersion string `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty"`
	// Kind of objects.
	Kind string `yaml:"kind,omitempty" json:"kind,omitempty"`
	// NameSelector used to subscribe on object by its name.
	NameSelector *NameSelector `yaml:"nameSelector,omitempty" json:"nameSelector,omitempty"`
	// NamespaceSelector used to subscribe on objects in namespaces.
	NamespaceSelector *NamespaceSelector `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	// LabelSelector used to subscribe on objects by matching their labels.
	LabelSelector *metav1.LabelSelector `yaml:"labelSelector,omitempty" json:"labelSelector,omitempty"`
	// FieldSelector used to subscribe on objects by matching specific fields (the list of fields is narrow, see shell-operator documentation).
	FieldSelector *FieldSelector `yaml:"fieldSelector,omitempty" json:"fieldSelector,omitempty"`
	// ExecuteHookOnEvents is true by default. Set to false if only snapshot update is needed.
	// *bool --> ExecuteHookOnEvents: [All events] || empty slice || nil
	ExecuteHookOnEvents *bool `yaml:"executeHookOnEvent,omitempty" json:"executeHookOnEvent,omitempty"`
	// ExecuteHookOnSynchronization is true by default. Set to false if only snapshot update is needed.
	// true || false
	ExecuteHookOnSynchronization *bool `yaml:"executeHookOnSynchronization,omitempty" json:"executeHookOnSynchronization,omitempty"`
	// WaitForSynchronization is true by default. Set to false if beforeHelm is not required this snapshot on start.
	// true || false
	WaitForSynchronization *bool `yaml:"waitForSynchronization,omitempty" json:"waitForSynchronization,omitempty"`
	// false by default
	// if JQ filter is empty - KeepFullObjectsInMemory always true
	KeepFullObjectsInMemory *bool `yaml:"keepFullObjectsInMemory,omitempty" json:"keepFullObjectsInMemory,omitempty"`
	// JQ filter to filter results from kubernetes objects
	JqFilter string `yaml:"jqFilter,omitempty" json:"jqFilter,omitempty"`

	AllowFailure            *bool  `yaml:"allowFailure,omitempty" json:"allowFailure,omitempty"`
	ResynchronizationPeriod string `yaml:"resynchronizationPeriod,omitempty" json:"resynchronizationPeriod,omitempty"`

	IncludeSnapshotsFrom []string `yaml:"includeSnapshotsFrom,omitempty" json:"includeSnapshotsFrom,omitempty"`

	Queue string `yaml:"queue,omitempty" json:"queue,omitempty"`
}

type Error struct {
	Message string `yaml:"message" json:"message"`
	Code    int    `yaml:"code,omitempty" json:"code,omitempty"`
}
