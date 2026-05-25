package framework

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// RegisterCRD makes a custom resource known to the fake cluster. After this
// call, objects of this kind can be supplied via KubeStateSet, listed via
// KubernetesResource, and used in KubernetesConfig snapshot bindings.
//
// Use it for CRDs that are not registered through a typed runtime.SchemeBuilder.
//
// Example:
//
//	hec.RegisterCRD("example.com", "v1alpha1", "Widget", true)
func (h *HookExecutionConfig) RegisterCRD(group, version, kind string, namespaced bool) {
	gvk := schema.GroupVersionKind{Group: group, Version: version, Kind: kind}
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	h.gvkToGVR[gvk] = gvr
	h.gvrToListKind[gvr] = kind + "List"

	// Make the GVK known to the unstructured scheme so the fake client can
	// list and watch it.
	registerUnstructuredGVK(h.unstructuredScheme, gvk)

	// Re-create the fake client so it picks up the new GVR mapping.
	h.resetCluster()

	_ = namespaced // namespaced is reserved for future scope-tracking; see snapshots.go
}
