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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/deckhouse/module-sdk/pkg"
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
	IntervalInSeconds uint8
	ProbeFunc         func(ctx context.Context, input *pkg.HookInput) error
}

func NewReadinessHookEM(cfg *ReadinessHookConfig) (*pkg.HookConfig, pkg.ReconcileFunc) {
	if cfg == nil {
		panic("empty readiness config")
	}

	return NewReadinessConfig(cfg), CheckModuleReadiness(cfg)
}

func NewReadinessConfig(cfg *ReadinessHookConfig) *pkg.HookConfig {
	if cfg.IntervalInSeconds == 0 {
		cfg.IntervalInSeconds = 15
	}

	return &pkg.HookConfig{
		Schedule: []pkg.ScheduleConfig{
			{
				Name:    cfg.ModuleName + "-moduleReadinessSchedule",
				Crontab: fmt.Sprintf("*/%d * * * * *", cfg.IntervalInSeconds),
			},
		},
	}
}

const (
	conditionStatusIsReady = "IsReady"
	modulePhaseReconciling = "Reconciling"
	modulePhaseReady       = "Ready"
)

func CheckModuleReadiness(cfg *ReadinessHookConfig) func(ctx context.Context, input *pkg.HookInput) error {
	if cfg.ModuleName == "" {
		panic("empty readiness module name")
	}

	if cfg.ProbeFunc == nil {
		cfg.ProbeFunc = func(_ context.Context, input *pkg.HookInput) error {
			input.Logger.Info("default probe function")

			return nil
		}
	}

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

		// Get conditions
		uConditions, ok, err := unstructured.NestedSlice(uModule.Object, "status", "conditions")
		if err != nil {
			return fmt.Errorf("nested slice: failed to get status.conditions: %w", err)
		}

		if !ok {
			return errors.New("can't find status.conditions")
		}

		if len(uConditions) == 0 {
			return errors.New("status.conditions is empty")
		}

		phase, ok, err := unstructured.NestedString(uModule.Object, "status", "phase")
		if err != nil {
			return fmt.Errorf("nested string: failed to get status.phase: %w", err)
		}

		if !ok {
			return errors.New("can't find status.phase")
		}

		if phase != modulePhaseReconciling && phase != modulePhaseReady {
			logger.Debug("waiting for sustainable phase", slog.String("phase", phase))

			return nil
		}

		// Run probe and get status
		probeStatus := string(corev1.ConditionTrue)
		probeMessage := ""
		probePhase := modulePhaseReady
		probeReason := ""
		if err := cfg.ProbeFunc(ctx, input); err != nil {
			probeStatus = string(corev1.ConditionFalse)
			probeMessage = err.Error()
			probePhase = modulePhaseReconciling
			probeReason = "ReadinessProbeFailed"
		}

		// search IsReady condition
		condIdx := -1
		var cond map[string]interface{}

		for idx, rawCond := range uConditions {
			cond = rawCond.(map[string]interface{})
			if cond["type"].(string) == conditionStatusIsReady {
				condIdx = idx
				break
			}
		}

		if condIdx < 0 {
			cond["type"] = conditionStatusIsReady
			uConditions = append(uConditions, cond)
			condIdx = len(uConditions) - 1
		}

		cond["lastProbeTime"] = input.DC.GetClock().Now().Format("2006-01-02T15:04:05Z")

		if cond["message"] != probeMessage || probePhase != phase {
			if probeStatus != cond["status"] {
				cond["lastTransitionTime"] = input.DC.GetClock().Now().Format("2006-01-02T15:04:05Z")
			}

			// Update condition
			cond["status"] = probeStatus

			cond["message"] = probeMessage
			if probeMessage == "" {
				delete(cond, "message")
			}

			cond["reason"] = probeReason
			if probeReason == "" {
				delete(cond, "reason")
			}
		}

		uConditions[condIdx] = cond
		// Update module status phase
		phase = probePhase

		// Update module status phase
		if err := unstructured.SetNestedField(uModule.Object, phase, "status", "phase"); err != nil {
			return fmt.Errorf("failed to change status.phase: %w", err)
		}

		// Update module status conditions
		if err := unstructured.SetNestedSlice(uModule.Object, uConditions, "status", "conditions"); err != nil {
			return fmt.Errorf("failed to change status.conditions: %w", err)
		}

		if _, err = k8sClient.Dynamic().Resource(*GetModuleGVK()).UpdateStatus(ctx, uModule, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("update module resource: %w", err)
		}

		return nil
	}
}
