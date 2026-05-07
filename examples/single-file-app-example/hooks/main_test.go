package main_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/helpers"

	singlefileappexample "singlefileappexample"
)

// newAppInput builds a *pkg.ApplicationHookInput from an InputBuilder.
// The application hook input has a different shape than the regular
// HookInput, so we re-pack the relevant collectors.
func newAppInput(b *helpers.InputBuilder, settings pkg.ReadableValuesCollector) *pkg.ApplicationHookInput {
	in := b.Build()
	return &pkg.ApplicationHookInput{
		Snapshots: in.Snapshots,
		Values:    in.Values,
		Settings:  settings,
		Logger:    in.Logger,
	}
}

func TestHandle_GateClosed_LeavesValuesUntouched(t *testing.T) {
	settings := helpers.NewValuesFromJSON(`{"apiServersDiscovery":{"enabled":false}}`)

	b := helpers.NewInputBuilder(t)
	in := newAppInput(b, settings)

	require.NoError(t, singlefileappexample.Handle(context.Background(), in))
	assert.Empty(t, in.Values.GetPatches(), "no patches expected when gate is closed")
}

func TestHandle_GateOpen_WritesDiscoveredPodsIntoValues(t *testing.T) {
	settings := helpers.NewValuesFromJSON(`{"apiServersDiscovery":{"enabled":true}}`)

	b := helpers.NewInputBuilder(t).
		WithSnapshot(singlefileappexample.SnapshotKey,
			helpers.SnapshotFromObject("kube-apiserver-1"),
			helpers.SnapshotFromObject("kube-apiserver-2"),
		)
	in := newAppInput(b, settings)

	require.NoError(t, singlefileappexample.Handle(context.Background(), in))

	patches := in.Values.GetPatches()
	require.Len(t, patches, 1)
	assert.Equal(t, "/test/internal/apiServers", patches[0].Path)
	assert.JSONEq(t, `["kube-apiserver-1","kube-apiserver-2"]`, string(patches[0].Value))
}
