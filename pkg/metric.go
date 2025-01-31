package pkg

type MetricsCollector interface {
	Outputer

	// Inc increments the specified Counter metric
	Inc(name string, labels map[string]string, opts ...Option)
	// Add adds custom value for the specified Counter metric
	Add(name string, value float64, labels map[string]string, opts ...Option)
	// Set specifies the custom value for the Gauge metric
	Set(name string, value float64, labels map[string]string, opts ...Option)
	// Expire marks metric's group as expired
	Expire(group string)
}

type MetricsCollectorOption interface {
	Apply(op MetricsCollectorOptionApplier)
}

type MetricsCollectorOptionApplier interface {
	WithDefaultGroup(group string)
}

type Option interface {
	Apply(op Operation)
}

type Operation interface {
	WithGroup(group string)
}
