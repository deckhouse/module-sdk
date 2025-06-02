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

//nolint:unused
package objectpatch

import (
	"time"

	"github.com/deckhouse/module-sdk/pkg"
)

type Option func(optionsApplier pkg.RegistryOptionApplier)

func (opt Option) Apply(o pkg.RegistryOptionApplier) {
	opt(o)
}

// WithCA use custom CA certificate
func WithCA(ca string) Option {
	return func(optionsApplier pkg.RegistryOptionApplier) {
		optionsApplier.WithCA(ca)
	}
}

// WithInsecureSchema use http schema instead of https
func WithInsecureSchema(insecure bool) Option {
	return func(optionsApplier pkg.RegistryOptionApplier) {
		optionsApplier.WithInsecureSchema(insecure)
	}
}

// WithAuth use docker config base64 as authConfig
// if dockerCfg is empty - will use client without auth
func WithAuth(dockerCfg string) Option {
	return func(optionsApplier pkg.RegistryOptionApplier) {
		optionsApplier.WithAuth(dockerCfg)
	}
}

// WithUserAgent adds ua string to the User-Agent header
func WithUserAgent(ua string) Option {
	return func(optionsApplier pkg.RegistryOptionApplier) {
		optionsApplier.WithUserAgent(ua)
	}
}

// WithTimeout limit and request to a registry with a timeout
// default timeout is 30 seconds
func WithTimeout(t time.Duration) Option {
	return func(optionsApplier pkg.RegistryOptionApplier) {
		optionsApplier.WithTimeout(t)
	}
}

type registryOptions struct {
	ca          string
	useHTTP     bool
	withoutAuth bool
	dockerCfg   string
	userAgent   string
	timeout     time.Duration
}

func (opts *registryOptions) WithCA(ca string) {
	opts.ca = ca
}

func (opts *registryOptions) WithInsecureSchema(insecure bool) {
	opts.useHTTP = insecure
}

func (opts *registryOptions) WithAuth(dockerCfg string) {
	opts.dockerCfg = dockerCfg
	if dockerCfg == "" {
		opts.withoutAuth = true
	}
}

func (opts *registryOptions) WithUserAgent(ua string) {
	opts.userAgent = ua
}

func (opts *registryOptions) WithTimeout(timeout time.Duration) {
	opts.timeout = timeout
}
