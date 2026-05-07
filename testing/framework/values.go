package framework

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
	k8syaml "sigs.k8s.io/yaml"

	"github.com/deckhouse/module-sdk/internal/testutils"
	patchablevalues "github.com/deckhouse/module-sdk/pkg/patchable-values"
	sdkutils "github.com/deckhouse/module-sdk/pkg/utils"
)

// valuesStore is a thin wrapper around a values map. It keeps both the
// canonical map[string]any and a derived JSON representation (regenerated
// lazily after mutations).
type valuesStore struct {
	values map[string]any
}

// newValuesStore parses initial values from a JSON or YAML string.
// Empty strings are treated as "{}".
func newValuesStore(raw string) (*valuesStore, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = "{}"
	}
	var v map[string]any
	if err := k8syaml.Unmarshal([]byte(raw), &v); err != nil {
		return nil, fmt.Errorf("unmarshal values: %w", err)
	}
	if v == nil {
		v = map[string]any{}
	}
	return &valuesStore{values: v}, nil
}

// JSON returns the values as a JSON string.
func (v *valuesStore) JSON() []byte {
	data, err := json.Marshal(v.values)
	if err != nil {
		// the store always holds a JSON-serializable map[string]any
		panic(fmt.Errorf("marshal values store: %w", err))
	}
	return data
}

// Map returns a deep clone of the values map (suitable for passing to
// patchable-values, which mutates).
func (v *valuesStore) Map() map[string]any {
	data := v.JSON()
	out := map[string]any{}
	_ = json.Unmarshal(data, &out)
	return out
}

// Get reads a dotted path using gjson.
func (v *valuesStore) Get(path string) gjson.Result {
	return gjson.GetBytes(v.JSON(), path)
}

// SetByPath sets a value at a dotted path. Intermediate maps are created on demand.
func (v *valuesStore) SetByPath(path string, value any) {
	parts := splitPath(path)
	setNested(v.values, parts, value)
}

// SetByYAML parses a YAML string and sets it at path.
func (v *valuesStore) SetByYAML(path string, raw []byte) error {
	var parsed any
	if err := k8syaml.Unmarshal(raw, &parsed); err != nil {
		return fmt.Errorf("unmarshal yaml: %w", err)
	}
	v.SetByPath(path, parsed)
	return nil
}

// DeleteByPath removes a path (no-op if it doesn't exist).
func (v *valuesStore) DeleteByPath(path string) {
	parts := splitPath(path)
	deleteNested(v.values, parts)
}

// applyPatchOperations applies a list of utils.ValuesPatchOperation (RFC6902)
// to the current values and updates v.values in-place.
//
// Unlike a strict RFC6902 implementation, intermediate object paths missing
// from the document are created on the fly for "add" / "replace" operations.
// This matches the developer expectation of input.Values.Set("a.b.c", x)
// just working on an empty document.
func (v *valuesStore) applyPatchOperations(ops []*sdkutils.ValuesPatchOperation) error {
	if len(ops) == 0 {
		return nil
	}

	for _, op := range ops {
		if err := v.applyOne(op); err != nil {
			return err
		}
	}
	return nil
}

// applyOne handles a single JSON-Patch-like operation. We support add, replace,
// and remove with create-on-demand semantics.
func (v *valuesStore) applyOne(op *sdkutils.ValuesPatchOperation) error {
	parts, err := splitJSONPath(op.Path)
	if err != nil {
		return err
	}

	switch op.Op {
	case "add", "replace":
		var decoded any
		if len(op.Value) > 0 {
			if err := json.Unmarshal(op.Value, &decoded); err != nil {
				return fmt.Errorf("decode value at %q: %w", op.Path, err)
			}
		}
		setNested(v.values, parts, decoded)
		return nil
	case "remove":
		deleteNested(v.values, parts)
		return nil
	default:
		// Fall back to the testutils JSON-Patch applier for operations we
		// don't natively support (copy, move, test).
		patched, _, err := testutils.ApplyValuesPatch(
			testutils.Values(v.values),
			testutils.ValuesPatch{Operations: []*sdkutils.ValuesPatchOperation{op}},
			testutils.IgnoreNonExistentPaths,
		)
		if err != nil {
			return fmt.Errorf("apply json-patch: %w", err)
		}
		v.values = map[string]any(patched)
		return nil
	}
}

// splitJSONPath converts a JSON-Pointer path like "/a/b/c" into ["a","b","c"].
// JSON-Pointer escapes "~" → "~0", "/" → "~1".
func splitJSONPath(p string) ([]string, error) {
	if p == "" || p == "/" {
		return nil, nil
	}
	if p[0] != '/' {
		return nil, fmt.Errorf("invalid json-pointer %q (must start with '/')", p)
	}
	parts := strings.Split(p[1:], "/")
	for i, s := range parts {
		s = strings.ReplaceAll(s, "~1", "/")
		s = strings.ReplaceAll(s, "~0", "~")
		parts[i] = s
	}
	return parts, nil
}

// splitPath splits a dotted path into segments. Empty input returns nil.
func splitPath(p string) []string {
	if p == "" {
		return nil
	}
	return strings.Split(p, ".")
}

func setNested(m map[string]any, parts []string, value any) {
	if len(parts) == 0 {
		return
	}
	if len(parts) == 1 {
		m[parts[0]] = value
		return
	}
	next, ok := m[parts[0]].(map[string]any)
	if !ok {
		next = map[string]any{}
		m[parts[0]] = next
	}
	setNested(next, parts[1:], value)
}

func deleteNested(m map[string]any, parts []string) {
	if len(parts) == 0 {
		return
	}
	if len(parts) == 1 {
		delete(m, parts[0])
		return
	}
	next, ok := m[parts[0]].(map[string]any)
	if !ok {
		return
	}
	deleteNested(next, parts[1:])
}

// patchableSnapshot is a small helper to grab final patches from a
// patchablevalues.PatchableValues without exposing it directly to the user.
func patchableValuesFor(v *valuesStore) (*patchablevalues.PatchableValues, error) {
	return patchablevalues.NewPatchableValues(v.Map())
}

// ===== Public value accessors on HookExecutionConfig =====

// ValuesGet returns the current value at path (gjson dotted path).
func (h *HookExecutionConfig) ValuesGet(path string) gjson.Result {
	return h.values.Get(path)
}

// ConfigValuesGet returns the current config value at path.
func (h *HookExecutionConfig) ConfigValuesGet(path string) gjson.Result {
	return h.configValues.Get(path)
}

// ValuesSet sets a value at path. The value is written directly into the
// values store; it persists across RunHook calls.
func (h *HookExecutionConfig) ValuesSet(path string, value any) {
	h.values.SetByPath(path, value)
}

// ConfigValuesSet sets a config value at path.
func (h *HookExecutionConfig) ConfigValuesSet(path string, value any) {
	h.configValues.SetByPath(path, value)
}

// ValuesSetFromYaml parses YAML and sets the result at path.
func (h *HookExecutionConfig) ValuesSetFromYaml(path string, raw []byte) {
	if err := h.values.SetByYAML(path, raw); err != nil {
		h.t.Fatalf("framework: ValuesSetFromYaml: %v", err)
	}
}

// ConfigValuesSetFromYaml parses YAML and sets the result at path.
func (h *HookExecutionConfig) ConfigValuesSetFromYaml(path string, raw []byte) {
	if err := h.configValues.SetByYAML(path, raw); err != nil {
		h.t.Fatalf("framework: ConfigValuesSetFromYaml: %v", err)
	}
}

// ValuesDelete removes a value at path.
func (h *HookExecutionConfig) ValuesDelete(path string) { h.values.DeleteByPath(path) }

// ConfigValuesDelete removes a config value at path.
func (h *HookExecutionConfig) ConfigValuesDelete(path string) { h.configValues.DeleteByPath(path) }

// ValuesJSON returns the current values as a JSON string. Mostly useful for
// debugging or asserting full document state.
func (h *HookExecutionConfig) ValuesJSON() []byte { return h.values.JSON() }

// ConfigValuesJSON returns the current config values as a JSON string.
func (h *HookExecutionConfig) ConfigValuesJSON() []byte { return h.configValues.JSON() }
