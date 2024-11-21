package pkg

type MetricCollector interface {
	Outputer

	// Inc increments the specified Counter metric
	Inc(name string, labels map[string]string, opts ...MetricOption)
	// Add adds custom value for the specified Counter metric
	Add(name string, value float64, labels map[string]string, opts ...MetricOption)
	// Set specifies the custom value for the Gauge metric
	Set(name string, value float64, labels map[string]string, opts ...MetricOption)
	// Expire marks metric's group as expired
	Expire(group string)
}

type MetricOption interface {
	Apply(op MetricOperation)
}

type MetricOperation interface {
	WithGroup(group string)
}
