package controller

import (
	"context"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/internal/transport/file"
	"github.com/deckhouse/module-sdk/pkg"
)

type HookConfig struct {
	BindingContextPath string
	ValuesPath         string
	ConfigValuesPath   string

	// send to addon operator when config requested
	HookConfigPath string

	MetricsPath          string
	KubernetesPath       string
	ValuesJSONPath       string
	ConfigValuesJSONPath string

	CreateFilesByYourself bool
}

type ReadinessConfig struct {
	ModuleName        string
	IntervalInSeconds uint8
	ProbeFunc         func(ctx context.Context, input *pkg.HookInput) error
}

type Config struct {
	ModuleName      string
	HookConfig      *HookConfig
	ReadinessConfig *ReadinessConfig

	LogLevelRaw string
	LogLevel    log.Level
}

func (cfg *Config) GetFileConfig() *file.Config {
	return &file.Config{
		BindingContextPath: cfg.HookConfig.BindingContextPath,
		ValuesPath:         cfg.HookConfig.ValuesPath,
		ConfigValuesPath:   cfg.HookConfig.ConfigValuesPath,

		HookConfigPath: cfg.HookConfig.HookConfigPath,

		MetricsPath:          cfg.HookConfig.MetricsPath,
		KubernetesPath:       cfg.HookConfig.KubernetesPath,
		ValuesJSONPath:       cfg.HookConfig.ValuesJSONPath,
		ConfigValuesJSONPath: cfg.HookConfig.ConfigValuesJSONPath,

		CreateFilesByYourself: cfg.HookConfig.CreateFilesByYourself,
	}
}
