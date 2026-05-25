package framework

import (
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
)

// Option configures a HookExecutionConfig at construction time.
type Option interface {
	apply(*execOptions)
}

type optionFunc func(*execOptions)

func (f optionFunc) apply(o *execOptions) { f(o) }

// execOptions holds parsed framework options.
type execOptions struct {
	initValues             string
	initConfigValues       string
	valuesSchemaPath       string
	configValuesSchemaPath string
	extraSchemeBuilders    []runtime.SchemeBuilder
	crds                   []customCRD
}

type customCRD struct {
	group      string
	version    string
	kind       string
	namespaced bool
}

// WithInitialValues sets the initial Helm values JSON or YAML.
func WithInitialValues(v string) Option {
	return optionFunc(func(o *execOptions) {
		o.initValues = v
	})
}

// WithInitialConfigValues sets the initial module config values JSON or YAML.
func WithInitialConfigValues(v string) Option {
	return optionFunc(func(o *execOptions) {
		o.initConfigValues = v
	})
}

// WithSchemeBuilder registers an additional runtime.SchemeBuilder so that
// typed CRDs from your module can be used in YAML state and assertions.
func WithSchemeBuilder(builder runtime.SchemeBuilder) Option {
	return optionFunc(func(o *execOptions) {
		o.extraSchemeBuilders = append(o.extraSchemeBuilders, builder)
	})
}

// WithCRD registers a custom resource definition with the fake cluster so
// that resources of this kind can be created/listed via the dynamic client.
//
// Use this when your hook reads or writes CRs not registered through a
// runtime.SchemeBuilder.
func WithCRD(group, version, kind string, namespaced bool) Option {
	return optionFunc(func(o *execOptions) {
		o.crds = append(o.crds, customCRD{
			group: group, version: version, kind: kind, namespaced: namespaced,
		})
	})
}

// WithValuesSchema reads an OpenAPI v3 schema from the given file path,
// extracts a values document populated with all `default:` values, and
// uses it as the baseline for the framework's `Values`. Anything passed
// via WithInitialValues (or HookExecutionConfigInit's initValues) is then
// deep-merged on top, so test-supplied values override schema defaults.
//
// The schema may use the addon-operator `x-extend` extension to inherit
// `properties` / `required` from a sibling schema (typically
// config-values.yaml). See LoadOpenAPISchema for details.
//
// Construction fails the test (via testing.TB.Fatalf) if the file is
// missing or malformed. Use WithOpenAPIDir if you want missing files to
// be silently ignored.
func WithValuesSchema(path string) Option {
	return optionFunc(func(o *execOptions) {
		o.valuesSchemaPath = path
	})
}

// WithConfigValuesSchema does the same as WithValuesSchema for the
// module's config values schema (typically `openapi/config-values.yaml`).
func WithConfigValuesSchema(path string) Option {
	return optionFunc(func(o *execOptions) {
		o.configValuesSchemaPath = path
	})
}

// WithOpenAPIDir is a convenience wrapper that points the framework at a
// directory containing the standard module OpenAPI files:
//
//	<dir>/values.yaml
//	<dir>/config-values.yaml
//
// Either file may be absent. Whichever ones are present are loaded and
// their defaults are merged under any test-supplied values.
//
// Example:
//
//	hec := framework.NewHookExecutionConfig(t, cfg, handler,
//	    framework.WithOpenAPIDir("../openapi"),
//	    framework.WithInitialValues(`{"https": {"mode": "CertManager"}}`),
//	)
func WithOpenAPIDir(dir string) Option {
	return optionFunc(func(o *execOptions) {
		valuesPath := filepath.Join(dir, "values.yaml")
		if fileExists(valuesPath) {
			o.valuesSchemaPath = valuesPath
		}
		configPath := filepath.Join(dir, "config-values.yaml")
		if fileExists(configPath) {
			o.configValuesSchemaPath = configPath
		}
	})
}
