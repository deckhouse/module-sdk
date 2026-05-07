package framework

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

// KubeStateSet replaces the fake cluster state with the resources defined in
// the provided multi-document YAML manifest.
//
// Documents may be separated by '---'. Each document must include
// apiVersion, kind, metadata.name, and (for namespaced resources) metadata.namespace.
//
// All previously-stored objects are removed before the new state is applied,
// so a single test can call KubeStateSet multiple times to simulate state
// transitions.
//
// Snapshots used by RunHook are regenerated from the new cluster state and
// the hook's KubernetesConfig bindings.
func (h *HookExecutionConfig) KubeStateSet(yamlState string) {
	h.t.Helper()

	objs, err := parseYAMLDocuments(yamlState)
	if err != nil {
		h.t.Fatalf("framework: parse kube state: %v", err)
	}

	h.resetCluster()

	ctx := context.Background()
	for i := range objs {
		obj := &objs[i]
		gvr, err := h.gvrFor(obj.GetAPIVersion(), obj.GetKind())
		if err != nil {
			h.t.Fatalf("framework: cannot resolve GVR for %s/%s: %v", obj.GetAPIVersion(), obj.GetKind(), err)
		}

		ri := h.resourceInterface(gvr, obj.GetNamespace())
		_, err = ri.Create(ctx, obj, metav1.CreateOptions{})
		if err == nil {
			continue
		}
		if apierrors.IsAlreadyExists(err) {
			if _, uerr := ri.Update(ctx, obj, metav1.UpdateOptions{}); uerr != nil {
				h.t.Fatalf("framework: update %s %s/%s: %v", obj.GetKind(), obj.GetNamespace(), obj.GetName(), uerr)
			}
			continue
		}
		h.t.Fatalf("framework: create %s %s/%s: %v", obj.GetKind(), obj.GetNamespace(), obj.GetName(), err)
	}
}

// AddKubeObject appends one or more objects (multi-document YAML) to the fake
// cluster without resetting existing state.
func (h *HookExecutionConfig) AddKubeObject(yamlObject string) {
	h.t.Helper()
	objs, err := parseYAMLDocuments(yamlObject)
	if err != nil {
		h.t.Fatalf("framework: parse object: %v", err)
	}
	ctx := context.Background()
	for i := range objs {
		obj := &objs[i]
		gvr, err := h.gvrFor(obj.GetAPIVersion(), obj.GetKind())
		if err != nil {
			h.t.Fatalf("framework: cannot resolve GVR for %s/%s: %v", obj.GetAPIVersion(), obj.GetKind(), err)
		}
		_, err = h.resourceInterface(gvr, obj.GetNamespace()).Create(ctx, obj, metav1.CreateOptions{})
		if err != nil {
			h.t.Fatalf("framework: create %s %s/%s: %v", obj.GetKind(), obj.GetNamespace(), obj.GetName(), err)
		}
	}
}

// resetCluster wipes the fake client's tracker by rebuilding it with the same
// scheme and ListKind mapping.
func (h *HookExecutionConfig) resetCluster() {
	h.fakeClient = dynamicfake.NewSimpleDynamicClientWithCustomListKinds(h.unstructuredScheme, h.gvrToListKind)
}

// resourceInterface returns the namespaced or cluster-scoped resource client
// for a given GVR.
func (h *HookExecutionConfig) resourceInterface(gvr schema.GroupVersionResource, namespace string) dynamic.ResourceInterface {
	r := h.fakeClient.Resource(gvr)
	if namespace == "" {
		return r
	}
	return r.Namespace(namespace)
}

// parseYAMLDocuments splits a multi-document YAML string into Unstructured
// objects, ignoring empty documents.
func parseYAMLDocuments(in string) ([]unstructured.Unstructured, error) {
	in = strings.TrimSpace(in)
	if in == "" {
		return nil, nil
	}

	reader := yaml.NewYAMLOrJSONDecoder(strings.NewReader(in), 4096)

	var out []unstructured.Unstructured
	for {
		raw := map[string]any{}
		if err := reader.Decode(&raw); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decode yaml: %w", err)
		}
		if len(raw) == 0 {
			continue
		}
		out = append(out, unstructured.Unstructured{Object: raw})
	}
	return out, nil
}

// KubernetesResource returns a fake-cluster resource by kind, namespace, and
// name. Namespace can be empty for cluster-scoped resources. Returns nil if
// the resource is not found.
func (h *HookExecutionConfig) KubernetesResource(kind, namespace, name string) *unstructured.Unstructured {
	h.t.Helper()
	gvr, err := h.resolveGVRForKind(kind)
	if err != nil {
		h.t.Fatalf("framework: KubernetesResource: %v", err)
	}
	obj, err := h.resourceInterface(gvr, namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		h.t.Fatalf("framework: get %s/%s: %v", namespace, name, err)
	}
	return obj
}

// KubernetesGlobalResource returns a cluster-scoped resource by kind and name.
// Returns nil if not found.
func (h *HookExecutionConfig) KubernetesGlobalResource(kind, name string) *unstructured.Unstructured {
	return h.KubernetesResource(kind, "", name)
}
