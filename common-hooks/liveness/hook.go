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

package liveness

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func getModuleGVK() *schema.GroupVersionResource {
	// ModuleGVR GroupVersionResource
	return &schema.GroupVersionResource{
		Group:    "deckhouse.io",
		Version:  "v1alpha1",
		Resource: "modules",
	}
}

type LivenessHookConfig struct {
	ModuleName        string
	IntervalInMinutes int
	ProbeFunc         func(ctx context.Context, input *pkg.HookInput) error
}

func RegisterLivenessHookEM(cfg *LivenessHookConfig) bool {
	if cfg == nil {
		panic("empty config")
	}

	return registry.RegisterFunc(&pkg.HookConfig{
		Schedule: []pkg.ScheduleConfig{
			{
				Name:    "moduleLivenessSchedule",
				Crontab: fmt.Sprintf("*/%d * * * *", cfg.IntervalInMinutes),
			},
		},
	}, checkModuleLiveness(cfg))
}

func checkModuleLiveness(cfg *LivenessHookConfig) func(ctx context.Context, input *pkg.HookInput) error {
	return func(ctx context.Context, input *pkg.HookInput) error {
		logger := input.Logger.With(slog.String("module", cfg.ModuleName))

		logger.Info("check liveness")

		k8sClient, err := input.DC.GetK8sClient()
		if err != nil {
			return fmt.Errorf("get k8s client: %w", err)
		}

		uModule, err := k8sClient.Dynamic().Resource(*getModuleGVK()).Get(ctx, cfg.ModuleName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("get module resource: %w", err)
		}

		if uModule == nil {
			return errors.New("unstructured object is nil")
		}

		probeStatus := "True"
		probeMessage := "success"

		err = cfg.ProbeFunc(ctx, input)
		if err != nil {
			probeStatus = "False"
			probeMessage = fmt.Sprintf("probe failed: %s", err)
		}

		uConditions, ok, err := unstructured.NestedSlice(uModule.Object, "status", "conditions")
		if err != nil {
			return fmt.Errorf("nested slice: %w", err)
		}

		if !ok {
			return errors.New("can't find status.conditions")
		}

		for idx, rawCond := range uConditions {
			cond := rawCond.(map[string]interface{})
			if cond["type"].(string) == "IsReady" {
				cond["status"] = probeStatus
				cond["message"] = probeMessage

				uConditions[idx] = cond

				break
			}
		}

		err = unstructured.SetNestedField(uModule.Object, uConditions, "status", "conditions")
		if err != nil {
			return fmt.Errorf("failed to change status.conditions: %w", err)
		}

		if _, err = k8sClient.Dynamic().Resource(*getModuleGVK()).UpdateStatus(ctx, uModule, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("update module resource: %w", err)
		}

		return nil
	}
}
