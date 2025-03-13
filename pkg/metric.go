package pkg

type OutputMetricsCollector interface {
	MetricsCollector
	Outputer
}

type MetricsCollector interface {
	// Inc increments the specified Counter metric
	Inc(name string, labels map[string]string, opts ...MetricCollectorOption)
	// Add adds custom value for the specified Counter metric
	Add(name string, value float64, labels map[string]string, opts ...MetricCollectorOption)
	// Set specifies the custom value for the Gauge metric
	Set(name string, value float64, labels map[string]string, opts ...MetricCollectorOption)
	// Expire marks metric's group as expired
	Expire(group string)
}

type MetricCollectorOption interface {
	Apply(op MetricCollectorOptionApplier)
}

type MetricCollectorOptionApplier interface {
	WithGroup(group string)
}
