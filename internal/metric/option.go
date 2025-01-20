package metric

import (
	"github.com/deckhouse/module-sdk/pkg"
)

var _ pkg.MetricsCollectorOption = (Option)(nil)

type Option func(o pkg.MetricsCollectorOptionApplier)

func (opt Option) Apply(o pkg.MetricsCollectorOptionApplier) {
	opt(o)
}

func WithDefaultGroup(group string) Option {
	return func(o pkg.MetricsCollectorOptionApplier) {
		o.WithDefaultGroup(group)
	}
}
