package kubernetes

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ObjectAndFilterResults map[string]*ObjectAndFilterResult

// ByNamespaceAndName implements sort.Interface for []ObjectAndFilterResult
// based on Namespace and Name of Object field.
type ByNamespaceAndName []ObjectAndFilterResult

func (a ByNamespaceAndName) Len() int      { return len(a) }
func (a ByNamespaceAndName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type ObjectAndFilterResult struct {
	Object       interface{} `json:"object,omitempty"`
	FilterResult interface{} `json:"filterResult,omitempty"`
}

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
