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

	"github.com/deckhouse/module-sdk/pkg"
	admissionv1 "k8s.io/api/admission/v1"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	apidiscoveryv2 "k8s.io/api/apidiscovery/v2"
	appsv1 "k8s.io/api/apps/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	authorizationv1 "k8s.io/api/authorization/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	certificatesv1 "k8s.io/api/certificates/v1"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	eventsv1 "k8s.io/api/events/v1"
	flowcontrolv1 "k8s.io/api/flowcontrol/v1"
	networkingv1 "k8s.io/api/networking/v1"
	nodev1 "k8s.io/api/node/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ pkg.KubernetesClient = (*Client)(nil)

type Client struct {
	client.Client
	dynamicClient dynamic.Interface
}

var defaultBuilders = []runtime.SchemeBuilder{
	admissionv1.SchemeBuilder,
	admissionregv1.SchemeBuilder,
	apidiscoveryv2.SchemeBuilder,
	appsv1.SchemeBuilder,
	authenticationv1.SchemeBuilder,
	authorizationv1.SchemeBuilder,
	autoscalingv1.SchemeBuilder,
	autoscalingv2.SchemeBuilder,
	batchv1.SchemeBuilder,
	certificatesv1.SchemeBuilder,
	coordinationv1.SchemeBuilder,
	corev1.SchemeBuilder,
	discoveryv1.SchemeBuilder,
	eventsv1.SchemeBuilder,
	flowcontrolv1.SchemeBuilder,
	networkingv1.SchemeBuilder,
	nodev1.SchemeBuilder,
	policyv1.SchemeBuilder,
	rbacv1.SchemeBuilder,
	schedulingv1.SchemeBuilder,
	storagev1.SchemeBuilder,
}

func NewClient(opts ...pkg.KubernetesOption) (*Client, error) {
	scheme := runtime.NewScheme()

	cfg := &clientOptions{}
	for _, opt := range opts {
		opt.Apply(cfg)
	}

	for _, builder := range defaultBuilders {
		utilruntime.Must(builder.AddToScheme(scheme))
	}

	for _, builder := range cfg.schemeBuilders {
		utilruntime.Must(builder.AddToScheme(scheme))
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

	return &Client{
		Client:        kClient,
		dynamicClient: d,
	}, nil
}

func (k Client) Dynamic() dynamic.Interface {
	return k.dynamicClient
}
