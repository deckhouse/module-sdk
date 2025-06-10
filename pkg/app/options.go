package app

import (
	"context"

	"github.com/deckhouse/module-sdk/pkg"
)

// RunConfigOption is a function that modifies the configuration for Run.
type RunConfigOption func(c *config)

// WithLogLevel sets the log level for the application.
func WithReadiness(cfg *ReadinessConfig) RunConfigOption {
	if cfg == nil {
		return func(c *config) {
			c.ReadinessConfig = nil
		}
	}

	return func(c *config) {
		c.ReadinessConfig = &readinessConfig{
			ProbeFunc: cfg.ProbeFunc,
		}

		if c.ReadinessConfig.IntervalInSeconds == 0 {
			c.ReadinessConfig.IntervalInSeconds = cfg.IntervalInSeconds
		}
	}
}

type ReadinessConfig struct {
	IntervalInSeconds uint8
	ProbeFunc         func(ctx context.Context, input *pkg.HookInput) error
}
