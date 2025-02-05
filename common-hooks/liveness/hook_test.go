/*
Copyright 2025 Flant JSC

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

package liveness_test

import (
	"context"
	"testing"

	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/deckhouse/module-sdk/common-hooks/liveness"
	"github.com/deckhouse/module-sdk/pkg"
	mock "github.com/deckhouse/module-sdk/testing/mock"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_LivenessHookConfig(t *testing.T) {
	t.Run("config is valid", func(t *testing.T) {
		assert.NoError(t, liveness.GenSelfSignedTLSConfig(&liveness.LivenessHookConfig{}).Validate())
	})
}

func Test_CheckModuleLiveness(t *testing.T) {
	t.Run("successfull check", func(t *testing.T) {
		mc := minimock.NewController(t)
		defer mc.Cleanup(func() {})

		dc := mock.NewDependencyContainerMock(mc)

		resource := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"status": map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{
							"type":    "IsReady",
							"status":  "False",
							"message": "Module is not ready",
						},
					},
				},
			},
		}

		updatedResource := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"status": map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{
							"type":    "IsReady",
							"status":  "True",
							"message": "Module is ready",
						},
					},
				},
			},
		}

		resourceMock := mock.NewKubernetesNamespaceableResourceInterfaceMock(mc)
		resourceMock.GetMock.
			Expect(minimock.AnyContext, "stub", metav1.GetOptions{}).
			Return(resource, nil)
		resourceMock.UpdateStatusMock.
			Expect(minimock.AnyContext, updatedResource, metav1.UpdateOptions{}).
			Return(nil, nil)

		dynamicClientMock := mock.NewKubernetesDynamicClientMock(mc)
		dynamicClientMock.ResourceMock.
			Expect(*liveness.GetModuleGVK()).
			Return(resourceMock)

		k8sClientMock := mock.NewKubernetesClientMock(mc)
		k8sClientMock.DynamicMock.
			Return(dynamicClientMock)

		dc.GetK8sClientMock.
			Expect().
			Return(k8sClientMock, nil)

		input := &pkg.HookInput{
			DC:     dc,
			Logger: log.NewNop(),
		}

		config := &liveness.LivenessHookConfig{
			ModuleName:        "stub",
			IntervalInMinutes: 10,
			ProbeFunc: func(ctx context.Context, input *pkg.HookInput) error {
				return nil
			},
		}

		err := liveness.CheckModuleLiveness(config)(context.Background(), input)
		assert.NoError(t, err)
	})
}
