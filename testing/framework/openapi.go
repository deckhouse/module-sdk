package framework

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	k8syaml "sigs.k8s.io/yaml"
)

// OpenAPI helpers for the testing framework.
//
// Modules ship two OpenAPI v3 schemas in their `openapi/` directory:
//
//   - config-values.yaml — describes the user-controllable module config;
//   - values.yaml        — extends config-values with system fields
//     (typically `internal:` and `registry:`).
//
// In Deckhouse, addon-operator validates and applies defaults from these
// schemas before invoking a hook. Hook tests, however, run the handler
// directly with whatever values the test author passed. That makes it
// easy to forget that a real environment would have, for example,
// `https.mode = "Disabled"` set by default — and tests then drift from
// production behaviour.
//
// The helpers here close that gap. They:
//
//  1. read an OpenAPI v3 schema from a file path;
//  2. extract a values document populated with all `default:` values
//     declared by the schema (recursively, including objects and arrays);
//  3. merge values supplied by the test on top of those defaults so the
//     test's values override the schema-provided ones.
//
// The functions are intentionally lightweight: they manipulate the schema
// as a `map[string]any` and do not validate values. For full validation
// use the validators in `pkg/values/validation` (or the upstream
// addon-operator implementation).

// LoadOpenAPISchema reads an OpenAPI v3 schema from a YAML or JSON file
// and returns the parsed document.
//
// If the schema document declares the addon-operator x-extend extension,
// e.g.:
//
//	x-extend:
//	  schema: config-values.yaml
//
// LoadOpenAPISchema also loads the referenced schema (resolved relative
// to the current schema's directory) and merges it as a parent: the
// parent's `properties`, `patternProperties`, `definitions`, `required`
// and extensions are folded into the current schema, with the current
// schema winning on conflicts. This mirrors the behaviour of
// addon-operator's ExtendTransformer.
//
// LoadOpenAPISchema does not resolve `$ref`s.
func LoadOpenAPISchema(path string) (map[string]any, error) {
	return loadOpenAPISchemaWithStack(path, nil)
}

// loadOpenAPISchemaWithStack tracks already-visited paths to break
// pathological cycles in x-extend chains.
func loadOpenAPISchemaWithStack(path string, stack []string) (map[string]any, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("openapi: resolve %q: %w", path, err)
	}
	for _, prev := range stack {
		if prev == abs {
			return nil, fmt.Errorf("openapi: x-extend cycle detected at %q", abs)
		}
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("openapi: read %q: %w", abs, err)
	}

	var doc map[string]any
	if err := k8syaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("openapi: parse %q: %w", abs, err)
	}
	if doc == nil {
		doc = map[string]any{}
	}

	parentPath, ok := extractExtendSchemaPath(doc)
	if !ok {
		return doc, nil
	}

	parentResolved := parentPath
	if !filepath.IsAbs(parentResolved) {
		parentResolved = filepath.Join(filepath.Dir(abs), parentResolved)
	}

	parent, err := loadOpenAPISchemaWithStack(parentResolved, append(stack, abs))
	if err != nil {
		return nil, fmt.Errorf("openapi: load x-extend parent %q: %w", parentPath, err)
	}

	mergeSchemaWithParent(doc, parent)
	return doc, nil
}

// extractExtendSchemaPath reads the optional `x-extend.schema` value
// from a schema document and returns it.
func extractExtendSchemaPath(doc map[string]any) (string, bool) {
	raw, ok := doc["x-extend"]
	if !ok {
		return "", false
	}
	settings, ok := raw.(map[string]any)
	if !ok {
		return "", false
	}
	schemaPath, ok := settings["schema"].(string)
	if !ok || schemaPath == "" {
		return "", false
	}
	return schemaPath, true
}

// mergeSchemaWithParent folds the parent schema's properties/required/etc.
// into the current schema, mirroring addon-operator's ExtendTransformer.
// The current schema wins on conflicts.
func mergeSchemaWithParent(current, parent map[string]any) {
	current["properties"] = mergeSchemaMap(current["properties"], parent["properties"])
	current["patternProperties"] = mergeSchemaMap(current["patternProperties"], parent["patternProperties"])
	current["definitions"] = mergeSchemaMap(current["definitions"], parent["definitions"])
	current["required"] = mergeRequired(current["required"], parent["required"])

	if _, has := current["title"]; !has {
		if title, ok := parent["title"].(string); ok && title != "" {
			current["title"] = title
		}
	}
	if _, has := current["description"]; !has {
		if desc, ok := parent["description"].(string); ok && desc != "" {
			current["description"] = desc
		}
	}

	for k, v := range parent {
		if k == "properties" || k == "patternProperties" || k == "definitions" ||
			k == "required" || k == "title" || k == "description" || k == "x-extend" {
			continue
		}
		if _, has := current[k]; has {
			continue
		}
		// Only carry over OpenAPI extensions and a small set of known
		// schema-level keys. We deliberately don't override `type`,
		// `properties`, etc. that the current schema already declared.
		if isExtension(k) {
			current[k] = v
		}
	}
}

// mergeSchemaMap merges two map-shaped schema fields (e.g. `properties`).
// Keys present in `current` win.
func mergeSchemaMap(current, parent any) any {
	out := map[string]any{}
	if pm, ok := parent.(map[string]any); ok {
		for k, v := range pm {
			out[k] = v
		}
	}
	if cm, ok := current.(map[string]any); ok {
		for k, v := range cm {
			out[k] = v
		}
	}
	if len(out) == 0 {
		// Preserve "field absent" instead of writing back an empty map.
		if current == nil && parent == nil {
			return nil
		}
	}
	return out
}

// mergeRequired deduplicates two `required:` lists (parent first).
func mergeRequired(current, parent any) any {
	pSlice := toStringSlice(parent)
	cSlice := toStringSlice(current)

	if len(pSlice) == 0 && len(cSlice) == 0 {
		if current == nil && parent == nil {
			return nil
		}
		return []any{}
	}

	seen := make(map[string]struct{}, len(pSlice)+len(cSlice))
	out := make([]any, 0, len(pSlice)+len(cSlice))
	for _, name := range pSlice {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	for _, name := range cSlice {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func toStringSlice(v any) []string {
	switch s := v.(type) {
	case []string:
		return s
	case []any:
		out := make([]string, 0, len(s))
		for _, item := range s {
			if str, ok := item.(string); ok {
				out = append(out, str)
			}
		}
		return out
	default:
		return nil
	}
}

func isExtension(key string) bool {
	return len(key) > 2 && key[0] == 'x' && key[1] == '-'
}

// SchemaDefaults walks an OpenAPI schema (as returned by LoadOpenAPISchema)
// and produces a values document populated with the schema's `default:`
// values.
//
// The algorithm mirrors addon-operator's defaulting:
//
//   - A property's `default:` (if any) is used as the starting value.
//   - For object-typed properties, sub-properties' defaults are then
//     applied for any keys the explicit `default:` left empty.
//   - For arrays, items' defaults are applied to each existing item of
//     the explicit `default:` list (item entries themselves are never
//     synthesised from `items.default`).
//
// SchemaDefaults always returns a non-nil map (possibly empty).
func SchemaDefaults(schema map[string]any) map[string]any {
	out, _ := schemaDefaults(schema)
	if out == nil {
		return map[string]any{}
	}
	m, ok := out.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return m
}

// schemaDefaults returns (defaultValue, hasDefault). It is recursive and
// works on any subschema, not just the top-level object.
func schemaDefaults(schema map[string]any) (any, bool) {
	if schema == nil {
		return nil, false
	}

	rawDef, hasDef := schema["default"]
	var base any
	if hasDef {
		base = deepCopyJSONValue(rawDef)
	}

	t, _ := schema["type"].(string)
	_, hasProps := schema["properties"]
	if t == "" && hasProps {
		t = "object"
	}

	switch t {
	case "object":
		baseMap, _ := base.(map[string]any)
		if hasDef && base != nil && baseMap == nil {
			// The default is a non-object value (e.g. null) for an
			// object-typed property — leave it alone.
			return base, true
		}
		if baseMap == nil {
			baseMap = map[string]any{}
		}
		props, _ := schema["properties"].(map[string]any)
		for name, raw := range props {
			sub, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			subDefault, has := schemaDefaults(sub)
			if !has {
				continue
			}
			existing, exists := baseMap[name]
			if !exists {
				baseMap[name] = subDefault
				continue
			}
			// Both sides exist. If both are maps, the explicit
			// default's entries win on conflict; otherwise leave
			// the explicit default untouched.
			if eMap, eOk := existing.(map[string]any); eOk {
				if subMap, sOk := subDefault.(map[string]any); sOk {
					baseMap[name] = MergeValues(subMap, eMap)
				}
			}
		}
		if hasDef || len(baseMap) > 0 {
			return baseMap, true
		}
		return nil, false

	case "array":
		// We only recurse into existing array entries (those provided
		// by the explicit default). We never create new entries from
		// `items.default` alone.
		if !hasDef {
			return nil, false
		}
		list, ok := base.([]any)
		if !ok {
			return base, true
		}
		items, _ := schema["items"].(map[string]any)
		if items == nil {
			return list, true
		}
		for i, item := range list {
			itemMap, ok := item.(map[string]any)
			if !ok {
				continue
			}
			itemDefaults, has := schemaDefaults(items)
			if !has {
				continue
			}
			itemMapDefaults, ok := itemDefaults.(map[string]any)
			if !ok {
				continue
			}
			list[i] = MergeValues(itemMapDefaults, itemMap)
		}
		return list, true

	default:
		if hasDef {
			return base, true
		}
		return nil, false
	}
}

// MergeValues deep-merges override into base and returns the result.
//
// Object-typed values are merged property-by-property (recursing). For
// arrays and scalar values the override replaces the base entirely.
//
// Neither input is modified.
//
// MergeValues is the natural counterpart to SchemaDefaults: combine
// defaults extracted from a schema with values supplied by the test,
// letting the test's values win on every conflict.
func MergeValues(base, override map[string]any) map[string]any {
	out := deepCopyJSONMap(base)
	if out == nil {
		out = map[string]any{}
	}
	for k, v := range override {
		if existing, ok := out[k]; ok {
			if eMap, eOk := existing.(map[string]any); eOk {
				if vMap, vOk := v.(map[string]any); vOk {
					out[k] = MergeValues(eMap, vMap)
					continue
				}
			}
		}
		out[k] = deepCopyJSONValue(v)
	}
	return out
}

// deepCopyJSONValue returns a deep copy of a JSON-compatible value.
// Maps and slices are cloned recursively; other values are returned as-is
// (they are immutable in JSON semantics).
func deepCopyJSONValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		return deepCopyJSONMap(t)
	case []any:
		out := make([]any, len(t))
		for i, item := range t {
			out[i] = deepCopyJSONValue(item)
		}
		return out
	default:
		return v
	}
}

func deepCopyJSONMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = deepCopyJSONValue(v)
	}
	return out
}

// fileExists returns true if path refers to an existing regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false
		}
		return false
	}
	return !info.IsDir()
}
