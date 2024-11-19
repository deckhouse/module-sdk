package main

import (
	"fmt"

	env "github.com/caarlos0/env/v11"
	"github.com/deckhouse/deckhouse/pkg/log"
)

type HookConfig struct {
	HookConfigPath     string `env:"HOOK_CONFIG_PATH" envDefault:"tmp/hook_config.json"`
	BindingContextPath string `env:"BINDING_CONTEXT_PATH" envDefault:"tmp/binding_context.json"`

	MetricsPath          string `env:"METRICS_PATH" envDefault:"tmp/metrics.json"`
	KubernetesPath       string `env:"KUBERNETES_PATCH_PATH" envDefault:"tmp/kubernetes.json"`
	ConfigValuesJSONPath string `env:"CONFIG_VALUES_JSON_PATCH_PATH" envDefault:"tmp/config_values.json"`
	ValuesJSONPath       string `env:"VALUES_JSON_PATCH_PATH" envDefault:"tmp/values.json"`
}

func newHookConfig() *HookConfig {
	return &HookConfig{}
}

type Config struct {
	HookConfig *HookConfig

	LogLevelRaw string    `env:"LOG_LEVEL" envDefault:"INFO"`
	LogLevel    log.Level `env:"-"`
}

func NewConfig() *Config {
	return &Config{
		HookConfig: newHookConfig(),
	}
}

func (cfg *Config) Parse() error {
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
