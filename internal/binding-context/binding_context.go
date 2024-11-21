package bindingcontext

import (
	v1 "k8s.io/api/admission/v1"
	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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

// from code-modules
type BindingContext struct {
	// Type of binding context: [Event, Synchronization, Group, Schedule]
	Type KubeEventType `json:"type,omitempty"`
	// Binding is a related binding name.
	Binding string `json:"binding,omitempty"`
	// Snapshots contain all objects for all bindings.
	Snapshots map[string][]byte `json:"snapshots,omitempty"`

	// For “Event”-type binding context on Kubernetes event
	WatchEvent   WatchEventType `json:"watchEvent,omitempty"`
	Object       any            `json:"object,omitempty"`
	FilterResult any            `json:"filterResult,omitempty"`

	// For “Synchronization”-type binding context
	Objects []map[string]any `json:"objects,omitempty"`
}

// from shell operator
type NewBC struct {
	BindingContextMetadata `json:"metadata"`

	// name of a binding or a group or kubeEventType if binding has no 'name' field
	Binding string `json:"binding,omitempty"`
	// additional fields for 'kubernetes' binding
	Type       KubeEventType  `json:"type,omitempty"`
	WatchEvent WatchEventType `json:"watchEvent,omitempty"`

	Objects []map[string]any `json:"objects,omitempty"`

	Snapshots map[string]any `json:"snapshots,omitempty"`

	AdmissionReview  *v1.AdmissionReview      `json:"admissionReview,omitempty"`
	ConversionReview *apixv1.ConversionReview `json:"conversionReview,omitempty"`
	FromVersion      string                   `json:"fromVersion,omitempty"`
	ToVersion        string                   `json:"toVersion,omitempty"`
}

type BindingContextMetadata struct {
	Version             string      `json:"version,omitempty"`
	BindingType         BindingType `json:"bindingType,omitempty"`
	JqFilter            string      `json:"jqFilter,omitempty"`
	IncludeSnapshots    []string    `json:"includeSnapshots,omitempty"`
	IncludeAllSnapshots bool        `json:"includeAllSnapshots,omitempty"`
	Group               string      `json:"group,omitempty"`
}

func (bc BindingContext) IsSynchronization() bool {
	return bc.Binding == "kubernetes" && bc.Type == TypeSynchronization
}
