package hookinfolder_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/helpers"
	"github.com/deckhouse/module-sdk/testing/mock"

	subfolder "example-module/subfolder"
)

// recorder collects metrics emitted by the hook so the test can assert
// on the parameters of every call.
type recorder struct {
	addCalls []metricCall
	setCalls []metricCall
	incCalls []metricCall
}

type metricCall struct {
	name   string
	value  float64
	labels map[string]string
}

func newMetricsRecorder(t *testing.T) (pkg.MetricsCollector, *recorder) {
	r := &recorder{}
	m := mock.NewMetricsCollectorMock(t)

	m.AddMock.Set(func(name string, value float64, labels map[string]string, _ ...pkg.MetricCollectorOption) {
		r.addCalls = append(r.addCalls, metricCall{name: name, value: value, labels: labels})
	})
	m.SetMock.Set(func(name string, value float64, labels map[string]string, _ ...pkg.MetricCollectorOption) {
		r.setCalls = append(r.setCalls, metricCall{name: name, value: value, labels: labels})
	})
	m.IncMock.Set(func(name string, labels map[string]string, _ ...pkg.MetricCollectorOption) {
		r.incCalls = append(r.incCalls, metricCall{name: name, labels: labels})
	})

	return m, r
}

func TestHandlerHookMetricsCollector_EmitsAllThreeMetrics(t *testing.T) {
	collector, rec := newMetricsRecorder(t)

	in := helpers.NewInputBuilder(t).
		WithMetricsCollector(collector).
		Build()

	require.NoError(t, subfolder.HandlerHookMetricsCollector(context.Background(), in))

	require.Len(t, rec.addCalls, 1)
	assert.Equal(t, metricCall{name: "stub-add-metric", value: 1, labels: map[string]string{"node_found": "node_name"}}, rec.addCalls[0])

	require.Len(t, rec.setCalls, 1)
	assert.Equal(t, metricCall{name: "stub-set-metric", value: 1, labels: map[string]string{"node_found": "node_name"}}, rec.setCalls[0])

	require.Len(t, rec.incCalls, 1)
	assert.Equal(t, metricCall{name: "stub-inc-metric", labels: map[string]string{"node_found": "node_name"}}, rec.incCalls[0])
}
