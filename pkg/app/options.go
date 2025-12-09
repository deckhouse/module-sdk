package app

import (
	"context"

	"github.com/deckhouse/module-sdk/pkg"
	settingscheck "github.com/deckhouse/module-sdk/pkg/settings-check"
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

func WithSettingsCheck(probeFunc settingscheck.SettingsCheckFunc) RunConfigOption {
	if probeFunc == nil {
		return func(c *config) {
			c.SettingsCheckConfig = nil
		}
	}

	return func(c *config) {
		c.SettingsCheckConfig = &settingsCheckConfig{
			ProbeFunc: probeFunc,
		}
	}
}
