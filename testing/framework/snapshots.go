package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/deckhouse/module-sdk/pkg"
	sdkjq "github.com/deckhouse/module-sdk/pkg/jq"
)

// generateSnapshots builds snapshots for every KubernetesConfig binding in
// the hook config based on the current fake-cluster state.
//
// For each binding it:
//   - resolves the GVR from APIVersion/Kind,
//   - lists matching objects (filtered by NameSelector / NamespaceSelector /
//     LabelSelector),
//   - applies the JqFilter (if any) to each matched object,
//   - stores the JSON result as a snapshot under the binding's Name.
func (h *HookExecutionConfig) generateSnapshots(ctx context.Context) (snapshotsMap, error) {
	out := snapshotsMap{}
	if h.hookConfig == nil {
		return out, nil
	}

	for _, b := range h.hookConfig.Kubernetes {
		// Allow empty APIVersion (defaults to "v1").
		apiVersion := b.APIVersion
		if apiVersion == "" {
			apiVersion = "v1"
		}

		gvr, err := h.gvrFor(apiVersion, b.Kind)
		if err != nil {
			return nil, fmt.Errorf("binding %q: %w", b.Name, err)
		}

		listOpts := metav1.ListOptions{}
		if b.LabelSelector != nil {
			sel, err := metav1.LabelSelectorAsSelector(b.LabelSelector)
			if err != nil {
				return nil, fmt.Errorf("binding %q: parse label selector: %w", b.Name, err)
			}
			listOpts.LabelSelector = sel.String()
		}

		// Determine which namespaces to inspect.
		namespaces, err := h.namespacesForBinding(ctx, &b)
		if err != nil {
			return nil, fmt.Errorf("binding %q: %w", b.Name, err)
		}

		var matched []unstructured.Unstructured
		for _, ns := range namespaces {
			list, err := h.resourceInterface(gvr, ns).List(ctx, listOpts)
			if err != nil {
				return nil, fmt.Errorf("binding %q: list %s in %q: %w", b.Name, gvr.Resource, ns, err)
			}
			for _, item := range list.Items {
				if !matchesNameSelector(item.GetName(), b.NameSelector) {
					continue
				}
				if !matchesFieldSelector(&item, b.FieldSelector) {
					continue
				}
				matched = append(matched, item)
			}
		}

		var compiledJQ *sdkjq.Query
		if jqExpr := strings.TrimSpace(b.JqFilter); jqExpr != "" {
			compiledJQ, err = sdkjq.NewQuery(jqExpr)
			if err != nil {
				return nil, fmt.Errorf("binding %q: compile jq filter %q: %w", b.Name, b.JqFilter, err)
			}
		}

		snaps := make([]pkg.Snapshot, 0, len(matched))
		for _, obj := range matched {
			snap, err := buildSnapshot(ctx, &obj, compiledJQ)
			if err != nil {
				return nil, fmt.Errorf("binding %q: build snapshot for %s/%s: %w", b.Name, obj.GetNamespace(), obj.GetName(), err)
			}
			snaps = append(snaps, snap)
		}
		out[b.Name] = snaps
	}
	return out, nil
}

// namespacesForBinding returns the list of namespaces to scan for a given
// KubernetesConfig binding. Returns [""] for cluster-scoped queries.
//
// Rules:
//   - NamespaceSelector == nil OR NameSelector matching no namespaces → []string{""}
//     (means "list everywhere"; the fake client treats Namespace("") as cluster-wide).
//   - NamespaceSelector.NameSelector populated → those exact namespaces.
//   - NamespaceSelector.LabelSelector populated → list namespaces, filter by label.
func (h *HookExecutionConfig) namespacesForBinding(ctx context.Context, b *pkg.KubernetesConfig) ([]string, error) {
	if b.NamespaceSelector == nil {
		return []string{""}, nil
	}
	if ns := b.NamespaceSelector.NameSelector; ns != nil && len(ns.MatchNames) > 0 {
		return ns.MatchNames, nil
	}
	if b.NamespaceSelector.LabelSelector != nil {
		gvr, err := h.gvrFor("v1", "Namespace")
		if err != nil {
			return nil, fmt.Errorf("resolve namespace gvr: %w", err)
		}
		sel, err := metav1.LabelSelectorAsSelector(b.NamespaceSelector.LabelSelector)
		if err != nil {
			return nil, fmt.Errorf("parse namespace label selector: %w", err)
		}
		list, err := h.resourceInterface(gvr, "").List(ctx, metav1.ListOptions{LabelSelector: sel.String()})
		if err != nil {
			return nil, fmt.Errorf("list namespaces: %w", err)
		}
		out := make([]string, 0, len(list.Items))
		for _, n := range list.Items {
			out = append(out, n.GetName())
		}
		return out, nil
	}
	return []string{""}, nil
}

func matchesNameSelector(name string, sel *pkg.NameSelector) bool {
	if sel == nil || len(sel.MatchNames) == 0 {
		return true
	}
	for _, n := range sel.MatchNames {
		if n == name {
			return true
		}
	}
	return false
}

func matchesFieldSelector(obj *unstructured.Unstructured, sel *pkg.FieldSelector) bool {
	if sel == nil || len(sel.MatchExpressions) == 0 {
		return true
	}
	for _, expr := range sel.MatchExpressions {
		val, _, _ := unstructured.NestedString(obj.Object, splitFieldPath(expr.Field)...)
		if !matchFieldOperator(val, expr.Operator, expr.Value) {
			return false
		}
	}
	return true
}

func splitFieldPath(field string) []string {
	field = strings.TrimPrefix(field, ".")
	return strings.Split(field, ".")
}

func matchFieldOperator(value, op, target string) bool {
	switch op {
	case "Equals", "=", "==":
		return value == target
	case "NotEquals", "!=":
		return value != target
	}
	return false
}

// buildSnapshot serialises an object (optionally through a JQ filter) into a
// pkg.Snapshot.
func buildSnapshot(ctx context.Context, obj *unstructured.Unstructured, q *sdkjq.Query) (pkg.Snapshot, error) {
	rawJSON, err := json.Marshal(obj.UnstructuredContent())
	if err != nil {
		return nil, fmt.Errorf("marshal object: %w", err)
	}
	if q == nil {
		return rawSnapshot(rawJSON), nil
	}
	res, err := q.FilterStringObject(ctx, string(rawJSON))
	if err != nil {
		return nil, fmt.Errorf("apply jq: %w", err)
	}
	return rawSnapshot([]byte(res.String())), nil
}

// silenceUnusedLabels keeps the labels import alive (used by some helpers
// during future expansion).
var _ = labels.Everything
