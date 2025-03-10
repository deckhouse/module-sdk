package app

import (
	"fmt"

	env "github.com/caarlos0/env/v11"

	"github.com/deckhouse/deckhouse/pkg/log"
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

type config struct {
	HookConfig *hookConfig

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
