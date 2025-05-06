package app

import (
	"log/slog"
	"runtime/debug"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/internal/controller"
)

func Run(opts ...RunConfigOption) {
	logger := log.Default()

	defer func() {
		// recover from panic if one occurred. Set err to nil otherwise.
		if r := recover(); r != nil {
			panicLogger := logger.WithGroup("panic").
				With(slog.Any("error", r)).
				With(slog.String("stacktrace", string(debug.Stack())))
			panicLogger.Error("panic recover")
		}
	}()

	cfg := newConfig()
	err := cfg.Parse()
	if err != nil {
		panic(err)
	}

	logger.SetLevel(cfg.LogLevel)

	for _, opt := range opts {
		opt(cfg)
	}

	controller := controller.NewHookController(remapConfigToControllerConfig(cfg), logger.Named("hook-controller"))

	c := newCMD(controller)

	c.Execute()
}
