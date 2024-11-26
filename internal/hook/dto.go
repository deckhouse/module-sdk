package hook

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type GoHookMetadata struct {
	// batch hook ID
	ID uint `yaml:"id" json:"id"`
	// Name is a key in snapshots map.
	Name string `yaml:"name" json:"name"`
	// Name is path to hook folder.
	Path string `yaml:"path" json:"path"`
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

type HookConfigSettings struct {
	ExecutionMinInterval time.Duration `yaml:"executionMinInterval,omitempty" json:"executionMinInterval,omitempty"`
	ExecutionBurst       int           `yaml:"executionBurst,omitempty" json:"executionBurst,omitempty"`
	// EnableSchedulesOnStartup
	// set to true, if you need to run 'Schedule' hooks without waiting addon-operator readiness
	EnableSchedulesOnStartup *bool
}

type ScheduleConfig struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Crontab is a schedule config in crontab format. (5 or 6 fields)
	Crontab string `yaml:"crontab" json:"crontab"`
	// Group                string   `yaml:"group,omitempty" json:"group,omitempty"`
	// Queue                string   `yaml:"queue,omitempty" json:"queue,omitempty"`
}

type FilterResult any

type FilterFunc func(*unstructured.Unstructured) (FilterResult, error)

type NameSelector struct {
	MatchNames []string `json:"matchNames" yaml:"matchNames"`
}

type FieldSelectorRequirement struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value,omitempty"`
}

type FieldSelector struct {
	MatchExpressions []FieldSelectorRequirement `json:"matchExpressions" yaml:"matchExpressions"`
}

type NamespaceSelector struct {
	NameSelector  *NameSelector         `json:"nameSelector,omitempty" yaml:"nameSelector,omitempty"`
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty" yaml:"labelSelector,omitempty"`
}

type KubernetesConfig struct {
	// APIVersion of objects. "v1" is used if not set.
	APIVersion string `json:"apiVersion,omitempty"`
	// Kind of objects.
	Kind string `json:"kind,omitempty"`
	// NameSelector used to subscribe on object by its name.
	NameSelector *NameSelector `json:"nameSelector,omitempty"`
	// NamespaceSelector used to subscribe on objects in namespaces.
	NamespaceSelector *NamespaceSelector `json:"namespace,omitempty"`
	// LabelSelector used to subscribe on objects by matching their labels.
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
	// FieldSelector used to subscribe on objects by matching specific fields (the list of fields is narrow, see shell-operator documentation).
	FieldSelector *FieldSelector `json:"fieldSelector,omitempty"`
	// ExecuteHookOnEvents is true by default. Set to false if only snapshot update is needed.
	// *bool --> ExecuteHookOnEvents: [All events] || либо пустой массив либо ничего
	ExecuteHookOnEvents *bool `json:"executeHookOnEvent,omitempty"`
	// ExecuteHookOnSynchronization is true by default. Set to false if only snapshot update is needed.
	// true || false
	ExecuteHookOnSynchronization *bool `json:"executeHookOnSynchronization,omitempty"`
	// WaitForSynchronization is true by default. Set to false if beforeHelm is not required this snapshot on start.
	// true || false
	WaitForSynchronization *bool `json:"waitForSynchronization,omitempty"`
	// false
	KeepFullObjectsInMemory *bool `json:"keepFullObjectsInMemory,omitempty"`

	JqFilter string `json:"jqFilter,omitempty"`

	AllowFailure            *bool  `json:"allowFailure,omitempty"`
	ResynchronizationPeriod string `json:"resynchronizationPeriod,omitempty"`

	IncludeSnapshotsFrom []string `json:"includeSnapshotsFrom,omitempty"`

	Queue string `json:"queue,omitempty"`
	// Named like hook (get from upper)
	Group string `json:"group,omitempty"`
}
