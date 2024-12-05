package app

import (
	"log/slog"
	"runtime/debug"

	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/deckhouse/module-sdk/internal/controller"
	"github.com/deckhouse/module-sdk/internal/transport/file"
)

func Run() {
	cfg := newConfig()
	err := cfg.Parse()
	if err != nil {
		panic(err)
	}

	logger := log.NewLogger(log.Options{
		Level: cfg.LogLevel.Level(),
	})

	fConfig := &file.Config{
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

	controller := controller.NewHookController(fConfig, logger.Named("hook-controller"))

	c := newCMD(controller)

	defer func() {
		// recover from panic if one occurred. Set err to nil otherwise.
		if r := recover(); r != nil {
			panicLogger := logger.WithGroup("panic").
				With(slog.Any("error", r)).
				With(slog.String("stacktrace", string(debug.Stack())))
			panicLogger.Error("panic recover")
		}
	}()

	c.Execute()
}
