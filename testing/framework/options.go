package framework

import "k8s.io/apimachinery/pkg/runtime"

// Option configures a HookExecutionConfig at construction time.
type Option interface {
	apply(*execOptions)
}

type optionFunc func(*execOptions)

func (f optionFunc) apply(o *execOptions) { f(o) }

// execOptions holds parsed framework options.
type execOptions struct {
	initValues          string
	initConfigValues    string
	extraSchemeBuilders []runtime.SchemeBuilder
	crds                []customCRD
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
