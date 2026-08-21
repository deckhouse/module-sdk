package operation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/pkg"
)

// Regression test: WithGroup used to be declared on a value receiver, so the
// assignment landed in a copy and the group was silently dropped.
func Test_Operation_WithGroup_MutatesReceiver(t *testing.T) {
	op := &Operation{Name: "d8_example_metric", Action: "set"}

	op.WithGroup("example_group")

	assert.Equal(t, "example_group", op.Group)
}

// The option produced by the package-level WithGroup must reach the operation
// through the applier interface, exactly as the collector invokes it.
func Test_WithGroup_Option_AppliesThroughApplier(t *testing.T) {
	op := &Operation{Name: "d8_example_metric", Action: "set"}

	var applier pkg.MetricCollectorOptionApplier = op
	WithGroup("example_group").Apply(applier)

	assert.Equal(t, "example_group", op.Group)
}

// *Operation must keep satisfying the applier interface after the receiver
// change, otherwise the collector would fail to compile.
func Test_Operation_ImplementsApplier(t *testing.T) {
	var _ pkg.MetricCollectorOptionApplier = (*Operation)(nil)
}

func Test_MetricOperationsFromReader_RoundTrip(t *testing.T) {
	data := []byte(`{"name":"m1","group":"g1","action":"set","value":1}` + "\n" +
		`{"name":"m2","action":"add","value":2}` + "\n")

	ops, err := MetricOperationsFromBytes(data)
	require.NoError(t, err)
	require.Len(t, ops, 2)

	assert.Equal(t, "m1", ops[0].Name)
	assert.Equal(t, "g1", ops[0].Group)
	assert.Equal(t, "set", ops[0].Action)

	assert.Equal(t, "m2", ops[1].Name)
	assert.Empty(t, ops[1].Group)
	assert.Equal(t, "add", ops[1].Action)
}

func Test_MetricOperationsFromBytes_Empty(t *testing.T) {
	ops, err := MetricOperationsFromBytes(nil)
	require.NoError(t, err)
	assert.Empty(t, ops)
}
