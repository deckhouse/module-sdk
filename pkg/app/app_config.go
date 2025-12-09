package app

import (
	"context"
	"fmt"

	"github.com/caarlos0/env/v11"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/internal/controller"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/settingscheck"
)

type hookConfig struct {
	BindingContextPath string `env:"BINDING_CONTEXT_PATH" envDefault:"in/binding_context.json"`
	ValuesPath         string `env:"VALUES_PATH" envDefault:"in/values_path.json"`
	ConfigValuesPath   string `env:"CONFIG_VALUES_PATH" envDefault:"in/config_values_path.json"`

	// send to addon operator when config requested
	HookConfigPath string `env:"HOOK_CONFIG_PATH" envDefault:"out/hook_config.json"`

	MetricsPath          string `env:"METRICS_PATH" envDefault:"out/metrics.json"`
	KubernetesPath       string `env:"KUBERNETES_PATCH_PATH" envDefault:"out/kubernetes.json"`
	ValuesJSONPath       string `env:"VALUES_JSON_PATCH_PATH" envDefault:"out/values.json"`
	ConfigValuesJSONPath string `env:"CONFIG_VALUES_JSON_PATCH_PATH" envDefault:"out/config_values.json"`

	CreateFilesByYourself bool `env:"CREATE_FILES" envDefault:"false"`
}

func newHookConfig() *hookConfig {
	return &hookConfig{}
}

type readinessConfig struct {
	ModuleName        string
	IntervalInSeconds uint8 `env:"INTERVAL_IN_SECONDS"`
	// TODO: not implemented
	Threshold int
	ProbeFunc func(ctx context.Context, input *pkg.HookInput) error
}

type config struct {
	ModuleName      string `env:"MODULE_NAME" envDefault:"default-module"`
	HookConfig      *hookConfig
	ReadinessConfig *readinessConfig `envPrefix:"READINESS_"`
	SettingsCheck   settingscheck.Check

	LogLevelRaw string    `env:"LOG_LEVEL" envDefault:"FATAL"`
	LogLevel    log.Level `env:"-"`
}

func newConfig() *config {
	return &config{
		HookConfig: newHookConfig(),
	}
}

func (cfg *config) Parse() error {
	opts := env.Options{
		Prefix: "",
	}

	err := env.ParseWithOptions(cfg, opts)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	cfg.LogLevel = log.LogLevelFromStr(cfg.LogLevelRaw)

	return nil
}

func remapConfigToControllerConfig(input *config) *controller.Config {
	cfg := &controller.Config{
		ModuleName: input.ModuleName,
		HookConfig: &controller.HookConfig{
			BindingContextPath:    input.HookConfig.BindingContextPath,
			ValuesPath:            input.HookConfig.ValuesPath,
			ConfigValuesPath:      input.HookConfig.ConfigValuesPath,
			HookConfigPath:        input.HookConfig.HookConfigPath,
			MetricsPath:           input.HookConfig.MetricsPath,
			KubernetesPath:        input.HookConfig.KubernetesPath,
			ValuesJSONPath:        input.HookConfig.ValuesJSONPath,
			ConfigValuesJSONPath:  input.HookConfig.ConfigValuesJSONPath,
			CreateFilesByYourself: input.HookConfig.CreateFilesByYourself,
		},

		LogLevelRaw: input.LogLevelRaw,
		LogLevel:    input.LogLevel,
	}

	if input.ReadinessConfig != nil {
		cfg.ReadinessConfig = &controller.ReadinessConfig{
			ModuleName:        input.ModuleName,
			IntervalInSeconds: input.ReadinessConfig.IntervalInSeconds,
			ProbeFunc:         input.ReadinessConfig.ProbeFunc,
		}
	}

	if input.SettingsCheck != nil {
		cfg.SettingsCheck = input.SettingsCheck
	}

	return cfg
}
