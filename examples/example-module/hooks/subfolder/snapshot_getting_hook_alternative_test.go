package hookinfolder_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/testing/helpers"
	"github.com/deckhouse/module-sdk/testing/mock"

	subfolder "example-module/subfolder"
)

func TestHandlerHookSnapshotsAlt_LogsAllNodes(t *testing.T) {
	b := helpers.NewInputBuilder(t).
		WithSnapshot(subfolder.NodeInfoSnapshotName,
			helpers.SnapshotFromObject(nodeInfo("first-node", "1")),
			helpers.SnapshotFromObject(nodeInfo("second-node", "2")),
		).
		WithCapturedLogger()

	require.NoError(t, subfolder.HandlerHookSnapshotsAlt(context.Background(), b.Build()))

	logs := strings.Split(b.LogBuffer().String(), "\n")
	require.GreaterOrEqual(t, len(logs), 3)
	assert.Contains(t, logs[0], `"msg":"hello from snapshot alt hook"`)
	assert.Contains(t, logs[1], `"Name":"first-node"`)
	assert.Contains(t, logs[2], `"Name":"second-node"`)
}

func TestHandlerHookSnapshotsAlt_PropagatesUnmarshalError(t *testing.T) {
	failing := mock.NewSnapshotMock(t).UnmarshalToMock.Set(func(_ any) error {
		return errors.New("boom")
	})

	in := helpers.NewInputBuilder(t).
		WithSnapshot(subfolder.NodeInfoSnapshotName, failing).
		Build()

	err := subfolder.HandlerHookSnapshotsAlt(context.Background(), in)
	require.Error(t, err)
	assert.ErrorContains(t, err, "boom")
}
