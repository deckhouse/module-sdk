package pkg

import (
	metric "github.com/deckhouse/module-sdk/pkg/metric/operation"
)

type MetricCollector interface {
	Outputer

	// Inc increments the specified Counter metric
	Inc(name string, labels map[string]string, opts ...metric.Option)
	// Add adds custom value for the specified Counter metric
	Add(name string, value float64, labels map[string]string, opts ...metric.Option)
	// Set specifies the custom value for the Gauge metric
	Set(name string, value float64, labels map[string]string, opts ...metric.Option)
	// Expire marks metric's group as expired
	Expire(group string)
	// Expire marks metric's group as expired
	WithGroup(group string) MetricCollector
}
