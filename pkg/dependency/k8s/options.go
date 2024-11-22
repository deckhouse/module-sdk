/*
Copyright 2024 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package k8s

import (
	"github.com/deckhouse/module-sdk/pkg"
	"k8s.io/apimachinery/pkg/runtime"
)

type Option func(optionsApplier pkg.KubernetesOptionApplier)

func (opt Option) Apply(o pkg.KubernetesOptionApplier) {
	opt(o)
}

// WithSchemeBuilder allows to add custom scheme builder to the controller-runtime client
// All default kubernetes schemes (k8s.io/api) is already added by default
func WithSchemeBuilder(builder runtime.SchemeBuilder) Option {
	return func(optionsApplier pkg.KubernetesOptionApplier) {
		optionsApplier.WithSchemeBuilder(builder)
	}
}

type clientOptions struct {
	schemeBuilders []runtime.SchemeBuilder
}

func (opts *clientOptions) WithSchemeBuilder(builder runtime.SchemeBuilder) {
	opts.schemeBuilders = append(opts.schemeBuilders, builder)
}
