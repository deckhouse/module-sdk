package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	operation "github.com/deckhouse/module-sdk/pkg/metric/operation"
)

// Set(..., WithGroup("g")) must land the group on the collected operation.
// This is the end-to-end path a Go hook actually uses.
func Test_Collector_Set_WithGroup(t *testing.T) {
	mc := NewCollector()

	mc.Set("d8_example_metric", 1, nil, operation.WithGroup("example_group"))

	metrics := mc.CollectedMetrics()
	require.Len(t, metrics, 1)
	assert.Equal(t, "example_group", metrics[0].Group)
	assert.Equal(t, "set", metrics[0].Action)
}

func Test_Collector_Add_WithGroup(t *testing.T) {
	mc := NewCollector()

	mc.Add("d8_example_metric", 2, nil, operation.WithGroup("example_group"))

	metrics := mc.CollectedMetrics()
	require.Len(t, metrics, 1)
	assert.Equal(t, "example_group", metrics[0].Group)
	assert.Equal(t, "add", metrics[0].Action)
}

func Test_Collector_Inc_WithGroup(t *testing.T) {
	mc := NewCollector()

	mc.Inc("d8_example_metric", nil, operation.WithGroup("example_group"))

	metrics := mc.CollectedMetrics()
	require.Len(t, metrics, 1)
	assert.Equal(t, "example_group", metrics[0].Group)
}

// WithGroup must override the collector's default group.
func Test_Collector_WithGroup_OverridesDefaultGroup(t *testing.T) {
	mc := NewCollector(WithDefaultGroup("default_group"))

	mc.Set("d8_example_metric", 1, nil, operation.WithGroup("explicit_group"))

	metrics := mc.CollectedMetrics()
	require.Len(t, metrics, 1)
	assert.Equal(t, "explicit_group", metrics[0].Group)
}

// Without an explicit group the default group is used.
func Test_Collector_DefaultGroup_Applied(t *testing.T) {
	mc := NewCollector(WithDefaultGroup("default_group"))

	mc.Set("d8_example_metric", 1, nil)

	metrics := mc.CollectedMetrics()
	require.Len(t, metrics, 1)
	assert.Equal(t, "default_group", metrics[0].Group)
}

// Applying WithGroup to one operation must not leak into a later one.
func Test_Collector_WithGroup_DoesNotLeakBetweenMetrics(t *testing.T) {
	mc := NewCollector()

	mc.Set("grouped", 1, nil, operation.WithGroup("g1"))
	mc.Set("ungrouped", 1, nil)

	metrics := mc.CollectedMetrics()
	require.Len(t, metrics, 2)
	assert.Equal(t, "g1", metrics[0].Group)
	assert.Empty(t, metrics[1].Group)
}

func Test_Collector_Expire(t *testing.T) {
	mc := NewCollector()

	mc.Expire("some_group")

	metrics := mc.CollectedMetrics()
	require.Len(t, metrics, 1)
	assert.Equal(t, "some_group", metrics[0].Group)
	assert.Equal(t, "expire", metrics[0].Action)
}
