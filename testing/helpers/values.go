package helpers

import (
	"encoding/json"
	"fmt"

	k8syaml "sigs.k8s.io/yaml"

	"github.com/deckhouse/module-sdk/pkg"
	patchablevalues "github.com/deckhouse/module-sdk/pkg/patchable-values"
)

// NewValues constructs a real pkg.PatchableValuesCollector seeded with the
// provided map. Use it when the hook under test should observe a working
// values store rather than a mock.
//
// Tests can then assert on the patches the hook produced via GetPatches().
func NewValues(initial map[string]any) pkg.PatchableValuesCollector {
	if initial == nil {
		initial = map[string]any{}
	}
	v, err := patchablevalues.NewPatchableValues(initial)
	if err != nil {
		panic(fmt.Errorf("helpers.NewValues: %w", err))
	}
	return v
}

// NewValuesFromJSON is like NewValues but parses a JSON document.
// Empty or whitespace-only input is treated as `{}`.
func NewValuesFromJSON(raw string) pkg.PatchableValuesCollector {
	return NewValues(parseJSONOrYAML(raw))
}

// NewValuesFromYAML is like NewValues but parses a YAML document.
// Empty or whitespace-only input is treated as `{}`.
func NewValuesFromYAML(raw string) pkg.PatchableValuesCollector {
	return NewValues(parseJSONOrYAML(raw))
}

func parseJSONOrYAML(raw string) map[string]any {
	if raw == "" {
		return map[string]any{}
	}
	var out map[string]any
	if err := k8syaml.Unmarshal([]byte(raw), &out); err != nil {
		panic(fmt.Errorf("helpers: parse values: %w", err))
	}
	if out == nil {
		out = map[string]any{}
	}
	return out
}

// MarshalValues returns the JSON encoding of the patch operations recorded
// on the values collector. It is a small convenience for snapshot-style
// assertions:
//
//	require.JSONEq(t, `[{"op":"add","path":"/foo","value":"bar"}]`,
//	    string(helpers.MarshalValues(values)))
func MarshalValues(v pkg.PatchableValuesCollector) []byte {
	out, err := json.Marshal(v.GetPatches())
	if err != nil {
		panic(fmt.Errorf("helpers.MarshalValues: %w", err))
	}
	return out
}
