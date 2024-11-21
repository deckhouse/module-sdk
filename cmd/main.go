package main

import (
	"io"

	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/deckhouse/module-sdk/internal/controller"
	"github.com/deckhouse/module-sdk/internal/transport/file"
	_ "github.com/deckhouse/module-sdk/registered-hooks"
)

func main() {
	cfg := NewConfig()
	err := cfg.Parse()
	if err != nil {
		panic(err)
	}

	logger := log.NewLogger(log.Options{
		Level:  cfg.LogLevel.Level(),
		Output: io.Discard,
	})

	fConfig := file.Config{
		BindingContextPath: cfg.HookConfig.BindingContextPath,
		ValuesPath:         cfg.HookConfig.ValuesPath,
		ConfigValuesPath:   cfg.HookConfig.ConfigValuesPath,

		HookConfigPath: cfg.HookConfig.HookConfigPath,

		MetricsPath:          cfg.HookConfig.MetricsPath,
		KubernetesPath:       cfg.HookConfig.KubernetesPath,
		ValuesJSONPath:       cfg.HookConfig.ValuesJSONPath,
		ConfigValuesJSONPath: cfg.HookConfig.ConfigValuesJSONPath,
	}

	controller := controller.NewHookController(fConfig, logger.Named("hook-controller"))

	c := NewCMD(controller)

	c.Execute()
}
