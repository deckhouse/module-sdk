package openapi

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// forkOnlyFields are the json keys this package adds on top of upstream
// JSONSchemaProps. Everything else must exist on both sides.
var forkOnlyFields = map[string]struct{}{
	"x-kubernetes-sensitive-data": {},
}

// TestForkCoversUpstreamFields is the guard that keeps the fork honest. JSONSchemaProps
// is copied by hand, so a k8s bump that adds a schema field — or retypes one — would
// otherwise make the installer silently strip or mis-decode that field on every CRD it
// applies, with no warning anywhere. If this fails after bumping
// k8s.io/apiextensions-apiserver, port the change into types.go rather than relaxing
// the test.
func TestForkCoversUpstreamFields(t *testing.T) {
	upstream := jsonFields(reflect.TypeOf(apiextensionsv1.JSONSchemaProps{}))
	fork := jsonFields(reflect.TypeOf(JSONSchemaProps{}))

	for tag, up := range upstream {
		f, ok := fork[tag]
		if !ok {
			t.Errorf("apiextensionsv1.JSONSchemaProps.%s (%q) is missing from the fork: the installer would strip it from every CRD", up.name, tag)

			continue
		}

		if want := forkType(up.typ); f.typ != want {
			t.Errorf("JSONSchemaProps.%s (%q) is %s upstream, so the fork must declare it %s, not %s: the installer would mis-decode it", up.name, tag, up.typ, want, f.typ)
		}

		if f.opts != up.opts {
			t.Errorf("JSONSchemaProps.%s (%q) has the json tag options %q upstream but %q in the fork: the installer would serialize it where the apiserver omits it, and every CRD would be updated on every reconcile", up.name, tag, up.opts, f.opts)
		}
	}

	for tag, f := range fork {
		if _, ok := upstream[tag]; ok {
			continue
		}

		if _, ok := forkOnlyFields[tag]; !ok {
			t.Errorf("JSONSchemaProps.%s (%q) exists in neither upstream nor forkOnlyFields: the apiserver will reject it as an unknown field", f.name, tag)
		}
	}
}

// TestForkCoversUpstreamUnions guards the three union types. Every one of their fields is
// json:"-" and carried by the hand-ported marshallers in marshal.go, so the tag-based
// guard above never sees them — they are the part of the fork most likely to drift
// unnoticed.
func TestForkCoversUpstreamUnions(t *testing.T) {
	for _, tc := range []struct{ fork, upstream reflect.Type }{
		{reflect.TypeOf(JSONSchemaPropsOrArray{}), reflect.TypeOf(apiextensionsv1.JSONSchemaPropsOrArray{})},
		{reflect.TypeOf(JSONSchemaPropsOrBool{}), reflect.TypeOf(apiextensionsv1.JSONSchemaPropsOrBool{})},
		{reflect.TypeOf(JSONSchemaPropsOrStringArray{}), reflect.TypeOf(apiextensionsv1.JSONSchemaPropsOrStringArray{})},
	} {
		t.Run(tc.fork.Name(), func(t *testing.T) {
			assert.Equal(t, forkStructFields(tc.upstream), structFields(tc.fork),
				"port the upstream change into types.go and marshal.go")
		})
	}
}

type fieldInfo struct{ name, typ, opts string }

// jsonFields maps the json tag name of every serialized field to its name, type and tag
// options. The options are part of the contract: a field that lost omitempty is serialized
// where the apiserver omits it, and the desired spec could never equal the stored one again.
func jsonFields(t reflect.Type) map[string]fieldInfo {
	out := make(map[string]fieldInfo, t.NumField())

	for i := range t.NumField() {
		field := t.Field(i)

		tag, opts, _ := strings.Cut(field.Tag.Get("json"), ",")
		if tag == "" || tag == "-" {
			continue
		}

		out[tag] = fieldInfo{name: field.Name, typ: field.Type.String(), opts: opts}
	}

	return out
}

// structFields maps every field name to its type, json:"-" ones included.
func structFields(t reflect.Type) map[string]string {
	out := make(map[string]string, t.NumField())

	for i := range t.NumField() {
		out[t.Field(i).Name] = t.Field(i).Type.String()
	}

	return out
}

// forkStructFields is structFields with every type rendered as the fork must declare it.
func forkStructFields(t reflect.Type) map[string]string {
	out := structFields(t)

	for name, typ := range out {
		out[name] = forkType(typ)
	}

	return out
}

// forkTypes are the upstream types this package retargets at itself. Every other type a
// field can hold — v1.JSON, v1.JSONSchemaURL, v1.ValidationRules — must stay upstream.
var forkTypes = []string{
	"JSONSchemaProps",
	"JSONSchemaPropsOrArray",
	"JSONSchemaPropsOrBool",
	"JSONSchemaPropsOrStringArray",
	"JSONSchemaDependencies",
	"JSONSchemaDefinitions",
}

// forkType renders an upstream field type as the fork must declare it. The mapping runs in
// this direction on purpose: erasing the package name on both sides instead would let a
// nested schema position left pointing at apiextensionsv1 compare equal — and that is the
// one drift these guards exist to catch, because Prune would then silently strip
// x-kubernetes-sensitive-data from every schema under it.
func forkType(upstream string) string {
	for _, name := range forkTypes {
		upstream = strings.ReplaceAll(upstream, "v1."+name, "openapi."+name)
	}

	return upstream
}

// roundTrip runs a schema through the fork the same way the installer does.
func roundTrip(t *testing.T, in map[string]any) map[string]any {
	t.Helper()

	out, err := Prune(in)
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
			"port": map[string]any{
				"type":       "integer",
				"minimum":    int64(1),
				"maximum":    int64(65535),
				"multipleOf": int64(2),
			},
			"ratio": map[string]any{
				"type":    "number",
				"maximum": 1.5,
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

	// The bounds are *float64 in Go but whole numbers on the wire, and the apiserver
	// returns them as int64. Encoding them as float64 would make the desired spec
	// permanently differ from the stored one and update every such CRD on every run.
	port := props["port"].(map[string]any)
	assert.Equal(t, int64(1), port["minimum"])
	assert.Equal(t, int64(65535), port["maximum"])
	assert.Equal(t, int64(2), port["multipleOf"])

	assert.Equal(t, 1.5, props["ratio"].(map[string]any)["maximum"], "a fractional bound stays a float")

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
