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
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/deckhouse/deckhouse/pkg/log"

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
	IntervalInSeconds int
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
	moduleReleaseIsReadyLabel = "modules.deckhouse.io/is-ready"
)

func CheckModuleReadiness(cfg *ReadinessHookConfig) func(ctx context.Context, input *pkg.HookInput) error {
	if cfg.ModuleName == "" {
		panic("empty readiness module name")
	}

	if cfg.ProbeFunc == nil {
		cfg.ProbeFunc = func(ctx context.Context, input *pkg.HookInput) error {
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

		labels, ok, err := unstructured.NestedStringMap(uModule.Object, "metadata", "labels")
		if err != nil {
			return fmt.Errorf("nested string map: failed to get metadata.labels %w", err)
		}

		if !ok {
			return errors.New("can't find metadata.labels")
		}

		isReady := labels[moduleReleaseIsReadyLabel]

		// Run probe
		err = cfg.ProbeFunc(ctx, input)
		if err != nil {
			logger.Warn("probe function failed", log.Err(err))
		}

		resultLabel := strconv.FormatBool(err == nil)
		if isReady == resultLabel {
			return nil
		}

		labels[moduleReleaseIsReadyLabel] = resultLabel

		// Update module status
		if err := unstructured.SetNestedStringMap(uModule.Object, labels, "metadata", "labels"); err != nil {
			return fmt.Errorf("failed to change metadata.labels: %w", err)
		}

		if _, err = k8sClient.Dynamic().Resource(*GetModuleGVK()).Update(ctx, uModule, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("update module resource: %w", err)
		}

		return nil
	}
}
