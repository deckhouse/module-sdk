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

package valuescheck

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/deckhouse/module-sdk/pkg"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func GetModuleGVR() *schema.GroupVersionResource {
	// ModuleGVR GroupVersionResource
	return &schema.GroupVersionResource{
		Group:    "deckhouse.io",
		Version:  "v1alpha1",
		Resource: "modules",
	}
}

type ValuesCheckHookConfig struct {
	ModuleName string
	ProbeFunc  func(ctx context.Context, input *pkg.HookInput) error
}

func NewValuesCheckHookEM(cfg *ValuesCheckHookConfig) (*pkg.HookConfig, pkg.ReconcileFunc) {
	if cfg == nil {
		panic("empty readiness config")
	}

	return NewValuesCheckConfig(cfg), ModuleValuesCheck(cfg)
}

func NewValuesCheckConfig(cfg *ValuesCheckHookConfig) *pkg.HookConfig {
	return &pkg.HookConfig{
		Schedule: []pkg.ScheduleConfig{
			{
				Name: cfg.ModuleName + "-moduleReadinessSchedule",
			},
		},
	}
}

const (
	conditionStatusIsReady = "IsReady"
	modulePhaseReconciling = "Reconciling"
	modulePhaseReady       = "Ready"
	modulePhaseHookError   = "Error"
)

func ModuleValuesCheck(cfg *ValuesCheckHookConfig) func(ctx context.Context, input *pkg.HookInput) error {
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

		uModule, err := k8sClient.Dynamic().Resource(*GetModuleGVR()).Get(ctx, cfg.ModuleName, metav1.GetOptions{})
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

		if phase != modulePhaseReconciling && phase != modulePhaseReady && phase != modulePhaseHookError {
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
			// if probe status changed - update time
			if probeStatus != cond["status"] {
				cond["lastTransitionTime"] = input.DC.GetClock().Now().Format("2006-01-02T15:04:05Z")
			}

			cond["status"] = probeStatus

			cond["message"] = probeMessage
			if probeMessage == "" {
				delete(cond, "message")
			}

			cond["reason"] = probeReason
			if probeReason == "" {
				delete(cond, "reason")
			}

			// Update module status phase
			phase = probePhase
		}

		uConditions[condIdx] = cond

		// creating patch
		patch := map[string]any{
			"status": map[string]any{
				"conditions": uConditions,
				"phase":      phase,
			},
		}

		input.PatchCollector.PatchWithMerge(
			patch,
			GetModuleGVR().GroupVersion().String(),
			"Module",
			"",
			cfg.ModuleName,
			objectpatch.WithSubresource("/status"),
		)

		return nil
	}
}
