package pkg

import (
	"context"
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

type GoHookMetadata struct {
	// Hook name
	Name string
	// Hook path
	Path string
}

type HookConfig struct {
	Metadata   GoHookMetadata
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
	// TODO: need helper to easy check jqfilter (for testing purpose)
	JqFilter string
	// Allow to fail hook
	AllowFailure *bool

	ResynchronizationPeriod string
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
