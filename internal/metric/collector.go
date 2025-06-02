package metric

import (
	"encoding/json"
	"fmt"
	"io"

	pointer "k8s.io/utils/ptr"

	"github.com/deckhouse/module-sdk/pkg"
	metric "github.com/deckhouse/module-sdk/pkg/metric/operation"
)

var _ pkg.MetricsCollector = (*Collector)(nil)
var _ MetricsCollectorOptionApplier = (*Collector)(nil)

type Collector struct {
	defaultGroup string

	metrics []metric.Operation
}

func NewCollector(opts ...MetricsCollectorOption) *Collector {
	c := &Collector{metrics: make([]metric.Operation, 0)}

	for _, opt := range opts {
		opt.Apply(c)
	}

	return c
}

func (mc *Collector) WithDefaultGroup(group string) {
	mc.defaultGroup = group
}

// Inc increments specified Counter metric
func (mc *Collector) Inc(name string, labels map[string]string, opts ...pkg.MetricCollectorOption) {
	mc.Add(name, 1, labels, opts...)
}

// Add adds custom value for Counter metric
func (mc *Collector) Add(name string, value float64, labels map[string]string, options ...pkg.MetricCollectorOption) {
	m := metric.Operation{
		Name:   name,
		Group:  mc.defaultGroup,
		Action: "add",
		Value:  pointer.To(value),
		Labels: labels,
	}

	for _, opt := range options {
		opt.Apply(&m)
	}

	mc.metrics = append(mc.metrics, m)
}

// Set specifies custom value for Gauge metric
func (mc *Collector) Set(name string, value float64, labels map[string]string, options ...pkg.MetricCollectorOption) {
	m := metric.Operation{
		Name:   name,
		Group:  mc.defaultGroup,
		Action: "set",
		Value:  pointer.To(value),
		Labels: labels,
	}

	for _, opt := range options {
		opt.Apply(&m)
	}

	mc.metrics = append(mc.metrics, m)
}

// Expire marks metric's group as expired
func (mc *Collector) Expire(group string) {
	if group == "" {
		group = mc.defaultGroup
	}

	mc.metrics = append(mc.metrics, metric.Operation{
		Group:  group,
		Action: "expire",
	})
}

func (mc *Collector) CollectedMetrics() []metric.Operation {
	return mc.metrics
}

func (mc *Collector) WriteOutput(w io.Writer) error {
	for _, m := range mc.metrics {
		err := json.NewEncoder(w).Encode(m)
		if err != nil {
			return fmt.Errorf("json marshall: %w", err)
		}
	}

	return nil
}
