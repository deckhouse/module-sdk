package openapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
)

// TestForkMarshalsLikeUpstream is the guard the reflection tests in types_test.go cannot
// be. Every field of the union types is json:"-", so the bytes they produce are decided by
// the hand-ported marshallers in marshal.go alone: a drift there — null instead of [], a
// lost Allows — changes what the installer applies to every CRD using that keyword while
// both field-set comparisons still pass. Only the bytes show it.
func TestForkMarshalsLikeUpstream(t *testing.T) {
	// x-kubernetes-sensitive-data is left out on purpose: the upstream type cannot hold it,
	// and TestRoundTripKeepsKnownFields already covers it
	raw := map[string]any{
		"type": "object",
		"properties": map[string]any{
			// items, single-schema form
			"tags": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			// items, tuple form
			"pair": map[string]any{"type": "array", "items": []any{
				map[string]any{"type": "string"},
				map[string]any{"type": "integer"},
			}},
			// additionalProperties in all three forms
			"labels": map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "string"}},
			"free":   map[string]any{"type": "object", "additionalProperties": true},
			"closed": map[string]any{"type": "object", "additionalProperties": false},
			"list":   map[string]any{"type": "array", "additionalItems": true},
			// dependencies in both forms
			"deps": map[string]any{"type": "object", "dependencies": map[string]any{
				"needsOther": []any{"other"},
				"needsShape": map[string]any{"type": "string"},
			}},
		},
	}

	fork := &JSONSchemaProps{}
	require.NoError(t, runtime.DefaultUnstructuredConverter.FromUnstructured(raw, fork))

	upstream := &apiextensionsv1.JSONSchemaProps{}
	require.NoError(t, runtime.DefaultUnstructuredConverter.FromUnstructured(raw, upstream))

	forkJSON, err := json.Marshal(fork)
	require.NoError(t, err)

	upstreamJSON, err := json.Marshal(upstream)
	require.NoError(t, err)

	assert.JSONEq(t, string(upstreamJSON), string(forkJSON),
		"port the upstream marshal.go change into this package's")
}
