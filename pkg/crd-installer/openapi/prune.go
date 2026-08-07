package openapi

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
)

// Prune returns the schema with every key that is not a field of JSONSchemaProps
// removed, at every nesting level.
//
// The result is encoded through JSON rather than through the reflection converter so
// that whole numbers come back as int64, exactly as they arrive from the apiserver.
// Reflection encodes maximum/minimum/multipleOf as float64 — the Go type of those
// fields — and a desired spec carrying float64 could never compare equal to the stored
// one, so every reconcile would issue an Update.
func Prune(raw map[string]any) (map[string]any, error) {
	props := &JSONSchemaProps{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw, props); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	data, err := json.Marshal(props)
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}

	clean := map[string]any{}
	if err := json.Unmarshal(data, &clean); err != nil {
		return nil, fmt.Errorf("decode pruned: %w", err)
	}

	return clean, nil
}
