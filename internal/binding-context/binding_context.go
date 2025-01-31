package bindingcontext

import (
	"encoding/json"
)

type BindingType string

const (
	Schedule             BindingType = "schedule"
	OnStartup            BindingType = "onStartup"
	OnKubernetesEvent    BindingType = "kubernetes"
	KubernetesConversion BindingType = "kubernetesCustomResourceConversion"
	KubernetesValidating BindingType = "kubernetesValidating"
	KubernetesMutating   BindingType = "kubernetesMutating"
)

type WatchEventType string

const (
	WatchEventAdded    WatchEventType = "Added"
	WatchEventModified WatchEventType = "Modified"
	WatchEventDeleted  WatchEventType = "Deleted"
)

type KubeEventType string

const (
	TypeSynchronization KubeEventType = "Synchronization"
	TypeEvent           KubeEventType = "Event"
	TypeSchedule        KubeEventType = "Schedule"
	TypeGroup           KubeEventType = "Group"
)

// from shell operator
type BindingContext struct {
	Metadata `json:"metadata"`
	// name of a binding or a group or kubeEventType if binding has no 'name' field
	Binding string `json:"binding,omitempty"`
	// additional fields for 'kubernetes' binding
	Type KubeEventType `json:"type,omitempty"`

	Snapshots map[string]ObjectAndFilterResults `json:"snapshots,omitempty"`

	// For “Event”-type binding context on Kubernetes event
	// WatchEvent WatchEventType `json:"watchEvent,omitempty"`
	// Object any `json:"object,omitempty"`
	// FilterResult any `json:"filterResult,omitempty"`

	// For “Synchronization”-type binding context
	// Objects []map[string]any `json:"objects,omitempty"`
}

type Metadata struct {
	Version             string      `json:"version,omitempty"`
	BindingType         BindingType `json:"bindingType,omitempty"`
	JqFilter            string      `json:"jqFilter,omitempty"`
	IncludeSnapshots    []string    `json:"includeSnapshots,omitempty"`
	IncludeAllSnapshots bool        `json:"includeAllSnapshots,omitempty"`
	Group               string      `json:"group,omitempty"`
}

type ObjectAndFilterResults []ObjectAndFilterResult

type ObjectAndFilterResult struct {
	Object       json.RawMessage `json:"object,omitempty"`
	FilterResult json.RawMessage `json:"filterResult,omitempty"`
}

func (bc BindingContext) IsSynchronization() bool {
	return bc.Binding == "kubernetes" && bc.Type == TypeSynchronization
}
