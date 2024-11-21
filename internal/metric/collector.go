package metric

import (
	"encoding/json"
	"fmt"
	"io"

	pointer "k8s.io/utils/ptr"

	metric "github.com/deckhouse/module-sdk/internal/metric/operation"
	"github.com/deckhouse/module-sdk/pkg"
	service "github.com/deckhouse/module-sdk/pkg"
)

var _ service.MetricCollector = (*Collector)(nil)

type Collector struct {
	defaultGroup string

	metrics []metric.Operation
}

func NewCollector() *Collector {
	return &Collector{metrics: make([]metric.Operation, 0)}
}

// Inc increments specified Counter metric
func (mc *Collector) Inc(name string, labels map[string]string, opts ...pkg.MetricOption) {
	mc.Add(name, 1, labels, opts...)
}

// Add adds custom value for Counter metric
func (mc *Collector) Add(name string, value float64, labels map[string]string, options ...pkg.MetricOption) {
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
func (mc *Collector) Set(name string, value float64, labels map[string]string, options ...pkg.MetricOption) {
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

func (mc *Collector) WriteOutput(w io.Writer) error {
	for _, m := range mc.metrics {
		err := json.NewEncoder(w).Encode(m)
		if err != nil {
			return fmt.Errorf("json marshall: %w", err)
		}
	}

	return nil
}