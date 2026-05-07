package main_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/testing/helpers"

	singlefileexample "singlefileexample"
)

func TestHandle_PopulatesValuesFromSnapshot(t *testing.T) {
	in := helpers.NewInputBuilder(t).
		WithSnapshot(singlefileexample.SnapshotKey,
			helpers.SnapshotFromObject("apiserver-1"),
			helpers.SnapshotFromObject("apiserver-2"),
		).
		Build()

	require.NoError(t, singlefileexample.Handle(context.Background(), in))

	patches := in.Values.GetPatches()
	require.Len(t, patches, 1, "expected exactly one Set call")

	op := patches[0]
	assert.Equal(t, "add", op.Op)
	assert.Equal(t, "/test/internal/apiServers", op.Path)
	assert.JSONEq(t, `["apiserver-1","apiserver-2"]`, string(op.Value))
}

func TestHandle_NoSnapshotsWritesEmptySlice(t *testing.T) {
	in := helpers.NewInputBuilder(t).Build()

	require.NoError(t, singlefileexample.Handle(context.Background(), in))

	patches := in.Values.GetPatches()
	require.Len(t, patches, 1)
	assert.JSONEq(t, `[]`, string(patches[0].Value))
}
