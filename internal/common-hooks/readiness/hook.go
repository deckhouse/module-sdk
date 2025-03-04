/*
Copyright 2022 Flant JSC

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

package readiness

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/deckhouse/module-sdk/pkg"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func GetModuleGVK() *schema.GroupVersionResource {
	// ModuleGVR GroupVersionResource
	return &schema.GroupVersionResource{
		Group:    "deckhouse.io",
		Version:  "v1alpha1",
		Resource: "modules",
	}
}

type ReadinessHookConfig struct {
	ModuleName        string
	IntervalInSeconds int
	ProbeFunc         func(ctx context.Context, input *pkg.HookInput) error
}

func NewReadinessHookEM(cfg *ReadinessHookConfig) (*pkg.HookConfig, pkg.ReconcileFunc) {
	if cfg == nil {
		panic("empty config")
	}

	return NewReadinessConfig(cfg), CheckModuleReadiness(cfg)
}

func NewReadinessConfig(cfg *ReadinessHookConfig) *pkg.HookConfig {
	if cfg.IntervalInSeconds == 0 {
		cfg.IntervalInSeconds = 1
	}

	return &pkg.HookConfig{
		Schedule: []pkg.ScheduleConfig{
			{
				Name:    "moduleReadinessSchedule",
				Crontab: fmt.Sprintf("*/%d * * * * *", cfg.IntervalInSeconds),
			},
		},
	}
}

const (
	conditionStatusIsReady = "IsReady"
)

func CheckModuleReadiness(cfg *ReadinessHookConfig) func(ctx context.Context, input *pkg.HookInput) error {
	return func(ctx context.Context, input *pkg.HookInput) error {
		logger := input.Logger.With(slog.String("module", cfg.ModuleName))
		logger.Info("check readiness")

		k8sClient, err := input.DC.GetK8sClient()
		if err != nil {
			return fmt.Errorf("get k8s client: %w", err)
		}

		uModule, err := k8sClient.Dynamic().Resource(*GetModuleGVK()).Get(ctx, cfg.ModuleName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("get module resource: %w", err)
		}

		if uModule == nil {
			return errors.New("unstructured object is nil")
		}

		// Run probe and get status
		probeStatus := string(corev1.ConditionTrue)
		probeMessage := "Module is ready"
		if err := cfg.ProbeFunc(ctx, input); err != nil {
			probeStatus = string(corev1.ConditionFalse)
			probeMessage = fmt.Sprintf("probe failed: %s", err)
		}

		// Get conditions
		uConditions, ok, err := unstructured.NestedSlice(uModule.Object, "status", "conditions")
		if err != nil {
			return fmt.Errorf("nested slice: %w", err)
		}

		if !ok {
			return errors.New("can't find status.conditions")
		}

		if len(uConditions) == 0 {
			return errors.New("status.conditions is empty")
		}

		// Update IsReady condition
		conditionUpdated := false
		for idx, rawCond := range uConditions {
			cond := rawCond.(map[string]interface{})
			if cond["type"].(string) == conditionStatusIsReady {
				cond["status"] = probeStatus
				cond["message"] = probeMessage
				uConditions[idx] = cond
				conditionUpdated = true
				break
			}
		}

		if !conditionUpdated {
			return fmt.Errorf("condition %s not found", conditionStatusIsReady)
		}

		// Update module status
		if err := unstructured.SetNestedSlice(uModule.Object, uConditions, "status", "conditions"); err != nil {
			return fmt.Errorf("failed to change status.conditions: %w", err)
		}

		if _, err = k8sClient.Dynamic().Resource(*GetModuleGVK()).UpdateStatus(ctx, uModule, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("update module resource: %w", err)
		}

		return nil
	}
}
