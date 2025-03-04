package app

import (
	"context"

	"github.com/deckhouse/module-sdk/pkg"
)

// RunConfigOption is a function that modifies the configuration for Run.
type RunConfigOption func(c *config)

// WithLogLevel sets the log level for the application.
func WithReadiness(cfg *ReadinessConfig) RunConfigOption {
	return func(c *config) {
		c.ReadinessConfig.ModuleName = cfg.ModuleName
		c.ReadinessConfig.IntervalInSeconds = cfg.IntervalInSeconds
		c.ReadinessConfig.ProbeFunc = cfg.ProbeFunc
	}
}

type ReadinessConfig struct {
	ModuleName        string
	IntervalInSeconds int
	ProbeFunc         func(ctx context.Context, input *pkg.HookInput) error
}
