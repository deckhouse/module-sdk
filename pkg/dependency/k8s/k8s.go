/*
Copyright 2021 Flant JSC

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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	client.Client
	Dynamic() dynamic.Interface
}

type K8sClient struct {
	client.Client
	dynamicClient dynamic.Interface
}

type clientOptions struct {
	addSchemes []func(s *runtime.Scheme) error
}

type Option func(cfg *clientOptions)

func NewClient(opts ...Option) (*K8sClient, error) {
	scheme := runtime.NewScheme()

	cfg := &clientOptions{}
	for _, opt := range opts {
		opt(cfg)
	}

	for _, schemeFn := range cfg.addSchemes {
		utilruntime.Must(schemeFn(scheme))
	}

	restConfig := ctrl.GetConfigOrDie()
	cOpts := client.Options{
		Scheme: scheme,
	}

	kClient, err := client.New(restConfig, cOpts)
	if err != nil {
		return nil, fmt.Errorf("create kubernetes client: %w", err)
	}

	d, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		panic(err.Error())
	}

	return &K8sClient{
		Client:        kClient,
		dynamicClient: d,
	}, nil
}

func (k K8sClient) Dynamic() dynamic.Interface {
	return k.dynamicClient
}

func WithAddToScheme(fn func(s *runtime.Scheme) error) Option {
	return func(cfg *clientOptions) {
		cfg.addSchemes = append(cfg.addSchemes, fn)
	}
}
