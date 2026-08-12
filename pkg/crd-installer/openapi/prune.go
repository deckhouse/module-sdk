package openapi

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/json"
)

// Prune returns the schema with every key that is not a field of JSONSchemaProps
// removed, at every nesting level.
//
// Everything goes through json rather than through the reflection converter, in both
// directions. Reflection is not an option here: it inlines an embedded field into the
// same map instead of applying json's depth rule, so it would decode every nested
// schema twice — once into JSONSchemaProps, once into the shadowed upstream field it
// then throws away — and with KUBE_PATCH_CONVERSION_DETECTOR set apimachinery compares
// the two decoders and klog.Fatalf's on the difference.
//
// The result comes back with whole numbers as int64, exactly as they arrive from the
// apiserver. Reflection encodes maximum/minimum/multipleOf as float64 — the Go type of
// those fields — and a desired spec carrying float64 could never compare equal to the
// stored one, so every reconcile would issue an Update.
func Prune(raw map[string]any) (map[string]any, error) {
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}

	// decoding into the type drops every key that is not one of its fields
	props := &JSONSchemaProps{}
	if err := json.Unmarshal(data, props); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	data, err = json.Marshal(props)
	if err != nil {
		return nil, fmt.Errorf("encode pruned: %w", err)
	}

	clean := map[string]any{}
	if err := json.Unmarshal(data, &clean); err != nil {
		return nil, fmt.Errorf("decode pruned: %w", err)
	}

	return clean, nil
}
