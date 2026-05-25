package framework

import (
	"context"
	"encoding/json"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/deckhouse/module-sdk/pkg"
	sdkjq "github.com/deckhouse/module-sdk/pkg/jq"
)

// applyPatchesToCluster applies the records collected from the hook to the
// fake cluster, mutating it in-place. It is called by RunHook after the
// hook handler finishes.
func (h *HookExecutionConfig) applyPatchesToCluster() error {
	if h.patchCollector == nil {
		return nil
	}

	ctx := context.Background()
	for _, p := range h.patchCollector.Records() {
		if err := h.applyPatch(ctx, p); err != nil {
			return fmt.Errorf("apply %s patch %s/%s: %w", p.Type, p.Namespace, p.Name, err)
		}
	}
	return nil
}

func (h *HookExecutionConfig) applyPatch(ctx context.Context, p RecordedPatch) error {
	switch p.Type {
	case PatchTypeCreate, PatchTypeCreateOrUpdate, PatchTypeCreateIfNotExists:
		return h.applyCreate(ctx, p)
	case PatchTypeDelete, PatchTypeDeleteInBackground, PatchTypeDeleteNonCascading:
		return h.applyDelete(ctx, p)
	case PatchTypeJSONPatch:
		return h.applyJSONPatch(ctx, p)
	case PatchTypeMergePatch:
		return h.applyMergePatch(ctx, p)
	case PatchTypeJQFilter:
		return h.applyJQFilter(ctx, p)
	}
	return fmt.Errorf("unknown patch type %q", p.Type)
}

func (h *HookExecutionConfig) applyCreate(ctx context.Context, p RecordedPatch) error {
	u, err := toUnstructured(p.Object)
	if err != nil {
		return fmt.Errorf("convert object: %w", err)
	}

	gvr, err := h.gvrFor(u.GetAPIVersion(), u.GetKind())
	if err != nil {
		return err
	}
	ri := h.resourceInterface(gvr, u.GetNamespace())

	switch p.Type {
	case PatchTypeCreate:
		_, err := ri.Create(ctx, u, metav1.CreateOptions{})
		return err
	case PatchTypeCreateIfNotExists:
		_, err := ri.Create(ctx, u, metav1.CreateOptions{})
		if err != nil && apierrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	case PatchTypeCreateOrUpdate:
		_, err := ri.Create(ctx, u, metav1.CreateOptions{})
		if err == nil {
			return nil
		}
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		// Pull current resourceVersion to allow Update.
		current, err := ri.Get(ctx, u.GetName(), metav1.GetOptions{})
		if err != nil {
			return err
		}
		u.SetResourceVersion(current.GetResourceVersion())
		_, err = ri.Update(ctx, u, metav1.UpdateOptions{})
		return err
	}
	return nil
}

func (h *HookExecutionConfig) applyDelete(ctx context.Context, p RecordedPatch) error {
	gvr, err := h.gvrFor(p.APIVersion, p.Kind)
	if err != nil {
		return err
	}
	err = h.resourceInterface(gvr, p.Namespace).Delete(ctx, p.Name, metav1.DeleteOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

func (h *HookExecutionConfig) applyJSONPatch(ctx context.Context, p RecordedPatch) error {
	gvr, err := h.gvrFor(p.APIVersion, p.Kind)
	if err != nil {
		return err
	}
	data, err := patchPayloadAsJSON(p.JSONPatch)
	if err != nil {
		return fmt.Errorf("marshal json patch: %w", err)
	}
	_, err = h.resourceInterface(gvr, p.Namespace).Patch(ctx, p.Name, types.JSONPatchType, data, metav1.PatchOptions{})
	if err != nil && apierrors.IsNotFound(err) && shouldIgnoreMissing(p.Options) {
		return nil
	}
	return err
}

func (h *HookExecutionConfig) applyMergePatch(ctx context.Context, p RecordedPatch) error {
	gvr, err := h.gvrFor(p.APIVersion, p.Kind)
	if err != nil {
		return err
	}
	data, err := patchPayloadAsJSON(p.MergePatch)
	if err != nil {
		return fmt.Errorf("marshal merge patch: %w", err)
	}
	_, err = h.resourceInterface(gvr, p.Namespace).Patch(ctx, p.Name, types.MergePatchType, data, metav1.PatchOptions{})
	if err != nil && apierrors.IsNotFound(err) && shouldIgnoreMissing(p.Options) {
		return nil
	}
	return err
}

func (h *HookExecutionConfig) applyJQFilter(ctx context.Context, p RecordedPatch) error {
	gvr, err := h.gvrFor(p.APIVersion, p.Kind)
	if err != nil {
		return err
	}
	ri := h.resourceInterface(gvr, p.Namespace)
	current, err := ri.Get(ctx, p.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) && shouldIgnoreMissing(p.Options) {
			return nil
		}
		return err
	}
	q, err := sdkjq.NewQuery(p.JQFilter)
	if err != nil {
		return fmt.Errorf("compile jq: %w", err)
	}
	res, err := q.FilterObject(ctx, current.UnstructuredContent())
	if err != nil {
		return fmt.Errorf("apply jq: %w", err)
	}
	var patched map[string]any
	if err := json.Unmarshal([]byte(res.String()), &patched); err != nil {
		return fmt.Errorf("decode jq result: %w", err)
	}
	current.Object = patched
	_, err = ri.Update(ctx, current, metav1.UpdateOptions{})
	return err
}

// patchPayloadAsJSON normalizes the patch payload to JSON bytes. The hook may
// pass a string, []byte, or any JSON-serializable value.
func patchPayloadAsJSON(payload any) ([]byte, error) {
	switch v := payload.(type) {
	case nil:
		return nil, fmt.Errorf("nil patch payload")
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return json.Marshal(v)
	}
}

func toUnstructured(obj any) (*unstructured.Unstructured, error) {
	switch v := obj.(type) {
	case *unstructured.Unstructured:
		return v, nil
	case unstructured.Unstructured:
		return &v, nil
	case map[string]any:
		return &unstructured.Unstructured{Object: v}, nil
	case runtime.Object:
		content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(v)
		if err != nil {
			return nil, err
		}
		return &unstructured.Unstructured{Object: content}, nil
	}
	// Fall back to round-tripping via JSON.
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	out := map[string]any{}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &unstructured.Unstructured{Object: out}, nil
}

// shouldIgnoreMissing inspects PatchCollectorOptions to detect WithIgnoreMissingObject(true).
// Because the option is opaque (an applier interface), we use a small helper applier to capture it.
func shouldIgnoreMissing(opts []pkg.PatchCollectorOption) bool {
	flag := &flagApplier{}
	for _, o := range opts {
		o.Apply(flag)
	}
	return flag.ignoreMissing
}

type flagApplier struct {
	subresource   string
	ignoreMissing bool
	ignoreHookErr bool
}

func (f *flagApplier) WithSubresource(s string)       { f.subresource = s }
func (f *flagApplier) WithIgnoreMissingObject(b bool) { f.ignoreMissing = b }
func (f *flagApplier) WithIgnoreHookError(b bool)     { f.ignoreHookErr = b }

// pkg import below is used for the flagApplier interface assertion.
// Keep this import here so the file is self-contained.
//

var _ = func() any { var _ pkg.PatchCollectorOptionApplier = (*flagApplier)(nil); return nil }()
