package openapi

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// forkOnlyFields are the json keys this package adds on top of upstream
// JSONSchemaProps. Everything else must exist on both sides.
var forkOnlyFields = map[string]struct{}{
	"x-kubernetes-sensitive-data": {},
}

// TestForkCoversUpstreamFields is the guard that keeps the fork honest. JSONSchemaProps
// is copied by hand, so a k8s bump that adds a schema field would otherwise make the
// installer silently strip that field from every CRD it applies, with no warning
// anywhere. If this fails after bumping k8s.io/apiextensions-apiserver, port the new
// field into types.go rather than relaxing the test.
func TestForkCoversUpstreamFields(t *testing.T) {
	upstream := jsonTags(reflect.TypeOf(apiextensionsv1.JSONSchemaProps{}))
	fork := jsonTags(reflect.TypeOf(JSONSchemaProps{}))

	for tag, name := range upstream {
		if _, ok := fork[tag]; !ok {
			t.Errorf("apiextensionsv1.JSONSchemaProps.%s (%q) is missing from the fork: the installer would strip it from every CRD", name, tag)
		}
	}

	for tag, name := range fork {
		if _, ok := upstream[tag]; ok {
			continue
		}

		if _, ok := forkOnlyFields[tag]; !ok {
			t.Errorf("JSONSchemaProps.%s (%q) exists in neither upstream nor forkOnlyFields: the apiserver will reject it as an unknown field", name, tag)
		}
	}
}

func jsonTags(t reflect.Type) map[string]string {
	out := make(map[string]string, t.NumField())

	for i := range t.NumField() {
		field := t.Field(i)

		tag, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if tag == "" || tag == "-" {
			continue
		}

		out[tag] = field.Name
	}

	return out
}

// roundTrip runs a schema through the fork the same way the installer does.
func roundTrip(t *testing.T, in map[string]any) map[string]any {
	t.Helper()

	props := &JSONSchemaProps{}
	require.NoError(t, runtime.DefaultUnstructuredConverter.FromUnstructured(in, props))

	out, err := runtime.DefaultUnstructuredConverter.ToUnstructured(props)
	require.NoError(t, err)

	return out
}

func TestRoundTripKeepsKnownFields(t *testing.T) {
	in := map[string]any{
		"type":                                 "object",
		"x-kubernetes-preserve-unknown-fields": true,
		"required":                             []any{"name"},
		"properties": map[string]any{
			"name": map[string]any{
				"type":      "string",
				"maxLength": int64(10),
				"enum":      []any{"a", "b"},
				"default":   "a",
			},
			"token": map[string]any{
				"type":                        "string",
				"x-kubernetes-sensitive-data": true,
			},
			// items in single-schema form
			"tags": map[string]any{
				"type":                   "array",
				"x-kubernetes-list-type": "set",
				"items": map[string]any{
					"type":      "string",
					"minLength": int64(1),
				},
			},
			// items in tuple form
			"pair": map[string]any{
				"type": "array",
				"items": []any{
					map[string]any{"type": "string"},
					map[string]any{"type": "integer"},
				},
			},
			// additionalProperties in schema form
			"labels": map[string]any{
				"type": "object",
				"additionalProperties": map[string]any{
					"type":                        "string",
					"x-kubernetes-sensitive-data": true,
				},
			},
			// additionalProperties in bool form
			"free": map[string]any{
				"type":                 "object",
				"additionalProperties": true,
			},
			"either": map[string]any{
				"allOf": []any{
					map[string]any{"type": "string"},
					map[string]any{"minLength": int64(2)},
				},
			},
		},
	}

	out := roundTrip(t, in)

	assert.Equal(t, "object", out["type"])
	assert.Equal(t, true, out["x-kubernetes-preserve-unknown-fields"])
	assert.Equal(t, []any{"name"}, out["required"])

	props := out["properties"].(map[string]any)

	name := props["name"].(map[string]any)
	// Numbers must come back as int64. A plain encoding/json round trip would yield
	// float64 here, which never compares equal to what the apiserver returns and would
	// make the installer issue an Update on every run.
	assert.Equal(t, int64(10), name["maxLength"])
	assert.Equal(t, []any{"a", "b"}, name["enum"])
	assert.Equal(t, "a", name["default"])

	assert.Equal(t, true, props["token"].(map[string]any)["x-kubernetes-sensitive-data"])

	tags := props["tags"].(map[string]any)
	assert.Equal(t, "set", tags["x-kubernetes-list-type"])
	assert.Equal(t, int64(1), tags["items"].(map[string]any)["minLength"])

	pair := props["pair"].(map[string]any)["items"].([]any)
	require.Len(t, pair, 2)
	assert.Equal(t, "integer", pair[1].(map[string]any)["type"])

	labels := props["labels"].(map[string]any)["additionalProperties"].(map[string]any)
	assert.Equal(t, true, labels["x-kubernetes-sensitive-data"],
		"the extension must survive inside additionalProperties too")

	assert.Equal(t, true, props["free"].(map[string]any)["additionalProperties"])

	either := props["either"].(map[string]any)["allOf"].([]any)
	require.Len(t, either, 2)
	assert.Equal(t, int64(2), either[1].(map[string]any)["minLength"])
}

func TestRoundTripDropsUnknownFields(t *testing.T) {
	in := map[string]any{
		"type":           "object",
		"x-doc-examples": []any{"root"},
		"x-description":  "root",
		"properties": map[string]any{
			"name": map[string]any{
				"type":                        "string",
				"x-doc-examples":              []any{"nested"},
				"x-doc-default":               "nested",
				"x-examples":                  []any{"nested"},
				"x-kubernetes-immutable":      true,
				"x-kubernetes-patch-strategy": "merge",
			},
			"tags": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":           "string",
					"x-doc-examples": []any{"in items"},
				},
			},
			"labels": map[string]any{
				"type": "object",
				"additionalProperties": map[string]any{
					"type":           "string",
					"x-doc-examples": []any{"in additionalProperties"},
				},
			},
		},
	}

	out := roundTrip(t, in)

	assert.NotContains(t, out, "x-doc-examples")
	assert.NotContains(t, out, "x-description")

	props := out["properties"].(map[string]any)

	name := props["name"].(map[string]any)
	assert.Equal(t, "string", name["type"], "known fields must survive alongside the dropped ones")

	for _, key := range []string{
		"x-doc-examples", "x-doc-default", "x-examples",
		// these two look official but are not in JSONSchemaProps, so a
		// x-kubernetes-* prefix rule would have let them through
		"x-kubernetes-immutable", "x-kubernetes-patch-strategy",
	} {
		assert.NotContains(t, name, key)
	}

	assert.NotContains(t, props["tags"].(map[string]any)["items"].(map[string]any), "x-doc-examples")
	assert.NotContains(t, props["labels"].(map[string]any)["additionalProperties"].(map[string]any), "x-doc-examples")
}
