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

package settingscheck

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/deckhouse/module-sdk/pkg"
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

type SettingsCheckHookConfig struct {
	ModuleName string
	ProbeFunc  SettingsCheckFunc
}

func NewSettingsCheckHookEM(cfg *SettingsCheckHookConfig) (*pkg.HookConfig, pkg.ReconcileFunc) {
	if cfg == nil {
		panic("empty readiness config")
	}

	return NewSettingsCheckConfig(cfg), SettingsCheck(cfg)
}

func NewSettingsCheckConfig(cfg *SettingsCheckHookConfig) *pkg.HookConfig {
	return &pkg.HookConfig{
		Schedule: []pkg.ScheduleConfig{
			{
				Name: cfg.ModuleName + "-moduleReadinessSchedule",
			},
		},
	}
}

// const (
// 	conditionStatusIsReady = "IsReady"
// 	modulePhaseReconciling = "Reconciling"
// 	modulePhaseReady       = "Ready"
// 	modulePhaseHookError   = "Error"
// )

type SettingsCheckHookResult struct {
	Allow   bool
	Message string
}

type SettingsCheckHookInput struct {
	Values pkg.OutputPatchableValuesCollector
	Logger pkg.Logger
	DC     pkg.DependencyContainer
}

type SettingsCheckFunc func(ctx context.Context, input *SettingsCheckHookInput) SettingsCheckHookResult

func SettingsCheck(cfg *SettingsCheckHookConfig) func(ctx context.Context, input *pkg.HookInput) error {
	if cfg.ModuleName == "" {
		panic("empty module name")
	}

	if cfg.ProbeFunc == nil {
		cfg.ProbeFunc = func(_ context.Context, input *SettingsCheckHookInput) SettingsCheckHookResult {
			input.Logger.Info("default probe function")

			return SettingsCheckHookResult{
				Allow: false,
			}
		}
	}

	return func(ctx context.Context, input *pkg.HookInput) error {
		logger := input.Logger.With(slog.String("module", cfg.ModuleName))
		logger.Info("check settings")

		// here we may check if the module is ready to serve requests

		probeInput := &SettingsCheckHookInput{
			Values: input.ConfigValues,
			Logger: logger,
			DC:     input.DC,
		}

		result := cfg.ProbeFunc(ctx, probeInput)

		if !result.Allow {
			return fmt.Errorf("settings check failed: %s", result.Message)
		}

		// if allow with message, warn about it
		if result.Message != "" {
			logger.Warn(result.Message)
		}

		return nil
	}
}
