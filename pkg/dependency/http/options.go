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

package http

import (
	"time"

	"github.com/deckhouse/module-sdk/pkg"
)

type Option func(optionsApplier pkg.HTTPOptionApplier)

func (opt Option) Apply(o pkg.HTTPOptionApplier) {
	opt(o)
}

// WithTimeout set custom timeout for http request. Default: 10 seconds
func WithTimeout(t time.Duration) Option {
	return func(optionsApplier pkg.HTTPOptionApplier) {
		optionsApplier.WithTimeout(t)
	}
}

// WithInsecureSkipVerify skip tls certificate validation
func WithInsecureSkipVerify() Option {
	return func(optionsApplier pkg.HTTPOptionApplier) {
		optionsApplier.WithInsecureSkipVerify()
	}
}

func WithAdditionalCACerts(certs [][]byte) Option {
	return func(optionsApplier pkg.HTTPOptionApplier) {
		optionsApplier.WithAdditionalCACerts(certs)
	}
}

func WithTLSServerName(name string) Option {
	return func(optionsApplier pkg.HTTPOptionApplier) {
		optionsApplier.WithTLSServerName(name)
	}
}

type httpOptions struct {
	timeout         time.Duration
	insecure        bool
	additionalTLSCA [][]byte
	tlsServerName   string
}

func (opts *httpOptions) WithTimeout(t time.Duration) {
	opts.timeout = t
}

func (opts *httpOptions) WithInsecureSkipVerify() {
	opts.insecure = true
}

func (opts *httpOptions) WithAdditionalCACerts(certs [][]byte) {
	opts.additionalTLSCA = append(opts.additionalTLSCA, certs...)
}

func (opts *httpOptions) WithTLSServerName(name string) {
	opts.tlsServerName = name
}
