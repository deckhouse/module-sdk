package pkg

type MetricsCollector interface {
	Outputer

	// Inc increments the specified Counter metric
	Inc(name string, labels map[string]string, opts ...MetricsOption)
	// Add adds custom value for the specified Counter metric
	Add(name string, value float64, labels map[string]string, opts ...MetricsOption)
	// Set specifies the custom value for the Gauge metric
	Set(name string, value float64, labels map[string]string, opts ...MetricsOption)
	// Expire marks metric's group as expired
	Expire(group string)
}

type MetricsOption interface {
	Apply(op MetricsOperation)
}

type MetricsOperation interface {
	WithGroup(group string)
}
