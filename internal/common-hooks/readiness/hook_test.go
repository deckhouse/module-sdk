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

package readiness_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/internal/common-hooks/readiness"
	"github.com/deckhouse/module-sdk/pkg"
	mock "github.com/deckhouse/module-sdk/testing/mock"
)

func Test_ReadinessHookConfig(t *testing.T) {
	t.Run("config is valid", func(t *testing.T) {
		assert.NoError(t, readiness.NewReadinessConfig(&readiness.ReadinessHookConfig{}).Validate())
	})
}

func Test_CheckModuleReadiness(t *testing.T) {
	t.Run("successful check", func(t *testing.T) {
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
					"phase": "Reconciling",
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
					"phase": "Ready",
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
			Expect(*readiness.GetModuleGVK()).
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

		config := &readiness.ReadinessHookConfig{
			ModuleName:        "stub",
			IntervalInSeconds: 10,
			ProbeFunc: func(_ context.Context, _ *pkg.HookInput) error {
				return nil
			},
		}

		err := readiness.CheckModuleReadiness(config)(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("k8s client error", func(t *testing.T) {
		mc := minimock.NewController(t)
		defer mc.Cleanup(func() {})

		dc := mock.NewDependencyContainerMock(mc)
		dc.GetK8sClientMock.
			Expect().
			Return(nil, fmt.Errorf("k8s client error"))

		input := &pkg.HookInput{
			DC:     dc,
			Logger: log.NewNop(),
		}

		config := &readiness.ReadinessHookConfig{
			ModuleName:        "stub",
			IntervalInSeconds: 10,
			ProbeFunc: func(_ context.Context, _ *pkg.HookInput) error {
				return nil
			},
		}

		err := readiness.CheckModuleReadiness(config)(context.Background(), input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "k8s client error")
	})

	t.Run("get resource error", func(t *testing.T) {
		mc := minimock.NewController(t)
		defer mc.Cleanup(func() {})

		dc := mock.NewDependencyContainerMock(mc)

		resourceMock := mock.NewKubernetesNamespaceableResourceInterfaceMock(mc)
		resourceMock.GetMock.
			Expect(minimock.AnyContext, "stub", metav1.GetOptions{}).
			Return(nil, fmt.Errorf("get error"))

		dynamicClientMock := mock.NewKubernetesDynamicClientMock(mc)
		dynamicClientMock.ResourceMock.
			Expect(*readiness.GetModuleGVK()).
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

		config := &readiness.ReadinessHookConfig{
			ModuleName:        "stub",
			IntervalInSeconds: 10,
			ProbeFunc: func(_ context.Context, _ *pkg.HookInput) error {
				return nil
			},
		}

		err := readiness.CheckModuleReadiness(config)(context.Background(), input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get error")
	})

	t.Run("update status error", func(t *testing.T) {
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
					"phase": "Reconciling",
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
					"phase": "Ready",
				},
			},
		}

		resourceMock := mock.NewKubernetesNamespaceableResourceInterfaceMock(mc)
		resourceMock.GetMock.
			Expect(minimock.AnyContext, "stub", metav1.GetOptions{}).
			Return(resource, nil)
		resourceMock.UpdateStatusMock.
			Expect(minimock.AnyContext, updatedResource, metav1.UpdateOptions{}).
			Return(nil, fmt.Errorf("update error"))

		dynamicClientMock := mock.NewKubernetesDynamicClientMock(mc)
		dynamicClientMock.ResourceMock.
			Expect(*readiness.GetModuleGVK()).
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

		config := &readiness.ReadinessHookConfig{
			ModuleName:        "stub",
			IntervalInSeconds: 10,
			ProbeFunc: func(_ context.Context, _ *pkg.HookInput) error {
				return nil
			},
		}

		err := readiness.CheckModuleReadiness(config)(context.Background(), input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "update error")
	})
}
