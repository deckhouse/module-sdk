package hookinfolder_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/helpers"
	"github.com/deckhouse/module-sdk/testing/mock"

	subfolder "example-module/subfolder"
)

func nodeInfo(name, uid string) subfolder.NodeInfo {
	return subfolder.NodeInfo{
		APIVersion: "v1",
		Kind:       "node",
		Metadata: subfolder.NodeInfoMetadata{
			Name:            name,
			ResourceVersion: "v1",
			UID:             uid,
		},
	}
}

func TestHandlerHookSnapshots_LogsAllNodes(t *testing.T) {
	b := helpers.NewInputBuilder(t).
		WithSnapshot(subfolder.NodeInfoSnapshotName,
			helpers.SnapshotFromObject(nodeInfo("first-node", "1")),
			helpers.SnapshotFromObject(nodeInfo("second-node", "2")),
		).
		WithCapturedLogger()

	require.NoError(t, subfolder.HandlerHookSnapshots(context.Background(), b.Build()))

	logs := strings.Split(b.LogBuffer().String(), "\n")
	require.GreaterOrEqual(t, len(logs), 3, "expected hello + two node-found logs")
	assert.Contains(t, logs[0], `"msg":"hello from snapshot hook"`)
	assert.Contains(t, logs[1], `"msg":"node found"`)
	assert.Contains(t, logs[1], `"Name":"first-node"`)
	assert.Contains(t, logs[2], `"Name":"second-node"`)
}

func TestHandlerHookSnapshots_PropagatesUnmarshalError(t *testing.T) {
	// We still want to assert on the propagation of the unmarshal error.
	// helpers.Snapshot* always unmarshals successfully (it's just JSON), so
	// here we use the existing minimock-generated SnapshotMock to inject an
	// error from UnmarshalTo.
	failing := mock.NewSnapshotMock(t).UnmarshalToMock.Set(func(_ any) error {
		return errors.New("boom")
	})

	in := helpers.NewInputBuilder(t).
		WithSnapshot(subfolder.NodeInfoSnapshotName, failing).
		Build()

	err := subfolder.HandlerHookSnapshots(context.Background(), in)
	require.Error(t, err)
	assert.ErrorContains(t, err, "boom")
}

func TestHandlerHookSnapshots_NoNodes(t *testing.T) {
	in := helpers.NewInputBuilder(t).Build()

	require.NoError(t, subfolder.HandlerHookSnapshots(context.Background(), in))

	// Sanity: with no snapshots, the input still satisfies the contract.
	assert.Empty(t, in.Snapshots.Get(subfolder.NodeInfoSnapshotName))
}

// Compile-time assertion that the InputBuilder produces a real *pkg.HookInput,
// so the hook handler signature is exercised.
var _ pkg.HookFunc[*pkg.HookInput] = subfolder.HandlerHookSnapshots
