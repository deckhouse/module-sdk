package metric

type MetricsCollectorOption interface {
	Apply(op MetricsCollectorOptionApplier)
}

type MetricsCollectorOptionApplier interface {
	WithDefaultGroup(group string)
}

var _ MetricsCollectorOption = (Option)(nil)

type Option func(o MetricsCollectorOptionApplier)

func (opt Option) Apply(o MetricsCollectorOptionApplier) {
	opt(o)
}

func WithDefaultGroup(group string) Option {
	return func(o MetricsCollectorOptionApplier) {
		o.WithDefaultGroup(group)
	}
}
