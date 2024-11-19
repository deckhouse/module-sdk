package config

import (
	"time"

	"github.com/deckhouse/module-sdk/pkg/kubernetes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type HookConfig struct {
	ConfigVersion string             `yaml:"configVersion" json:"configVersion"`
	Schedule      []ScheduleConfig   `yaml:"schedule,omitempty" json:"schedule,omitempty"`
	Kubernetes    []KubernetesConfig `yaml:"kubernetes,omitempty" json:"kubernetes,omitempty"`
	// OnStartup runs hook on module/global startup
	// Attention! During the startup you don't have snapshots available
	// use native KubeClient to fetch resources
	OnStartup         int                 `yaml:"onStartup,omitempty" json:"onStartup,omitempty"`
	OnBeforeHelm      int                 `yaml:"beforeHelm,omitempty" json:"beforeHelm,omitempty"`
	OnAfterHelm       int                 `yaml:"afterHelm,omitempty" json:"afterHelm,omitempty"`
	OnAfterDeleteHelm int                 `yaml:"afterDeleteHelm,omitempty" json:"afterDeleteHelm,omitempty"`
	Settings          *HookConfigSettings `yaml:"settings,omitempty" json:"settings,omitempty"`

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
	// AllowFailure         *bool    `yaml:"allowFailure,omitempty" json:"allowFailure,omitempty"`
	// IncludeSnapshotsFrom []string `yaml:"includeSnapshotsFrom,omitempty" json:"includeSnapshotsFrom,omitempty"`
}

type KubernetesConfig struct {
	// Name is a key in snapshots map.
	Name string `yaml:"name" json:"name"`
	// ApiVersion of objects. "v1" is used if not set.
	ApiVersion string `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty"`
	// Kind of objects.
	Kind string `yaml:"kind,omitempty" json:"kind,omitempty"`
	// NameSelector used to subscribe on object by its name.
	NameSelector *kubernetes.NameSelector `yaml:"nameSelector,omitempty" json:"nameSelector,omitempty"`
	// NamespaceSelector used to subscribe on objects in namespaces.
	NamespaceSelector *kubernetes.NamespaceSelector `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	// LabelSelector used to subscribe on objects by matching their labels.
	LabelSelector *v1.LabelSelector `yaml:"labelSelector,omitempty" json:"labelSelector,omitempty"`
	// FieldSelector used to subscribe on objects by matching specific fields (the list of fields is narrow, see shell-operator documentation).
	FieldSelector *kubernetes.FieldSelector `yaml:"fieldSelector,omitempty" json:"fieldSelector,omitempty"`
	// ExecuteHookOnEvents is true by default. Set to false if only snapshot update is needed.
	ExecuteHookOnEvents []string `yaml:"executeHookOnEvent,omitempty" json:"executeHookOnEvent,omitempty"`
	// ExecuteHookOnSynchronization is true by default. Set to false if only snapshot update is needed.
	ExecuteHookOnSynchronization *bool `yaml:"executeHookOnSynchronization,omitempty" json:"executeHookOnSynchronization,omitempty"`
	// WaitForSynchronization is true by default. Set to false if beforeHelm is not required this snapshot on start.
	WaitForSynchronization *bool `yaml:"waitForSynchronization,omitempty" json:"waitForSynchronization,omitempty"`

	// JqFilter                string   `yaml:"jqFilter,omitempty" json:"jqFilter,omitempty"`
	// Group                   string   `yaml:"group,omitempty" json:"group,omitempty"`
	// Queue                   string   `yaml:"queue,omitempty" json:"queue,omitempty"`
	// IncludeSnapshotsFrom    []string `yaml:"includeSnapshotsFrom,omitempty" json:"includeSnapshotsFrom,omitempty"`
	// KeepFullObjectsInMemory *bool    `yaml:"keepFullObjectsInMemory,omitempty" json:"keepFullObjectsInMemory,omitempty"`
}

type FilterResult interface{}

type FilterFunc func(*unstructured.Unstructured) (FilterResult, error)
