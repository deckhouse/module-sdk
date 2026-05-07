// Package helpers provides building blocks for hook unit tests.
//
// It complements the heavier-weight testing/framework package: where the
// framework spins up a fake Kubernetes cluster and exercises the full hook
// pipeline (snapshots → handler → patches), the helpers in this package
// are aimed at small, focused unit tests that mock just the dependencies
// the hook touches.
//
// Typical usage:
//
//	func TestMyHook(t *testing.T) {
//	    in := helpers.NewInputBuilder(t).
//	        WithSnapshot("nodes", helpers.SnapshotJSON(`{"name":"n1"}`)).
//	        WithValuesJSON(`{"my":{"field":"value"}}`).
//	        Build()
//
//	    err := MyHookHandler(context.Background(), in)
//	    require.NoError(t, err)
//	    require.Equal(t, "value", in.Values.Get("my.field").String())
//	}
//
// Helpers are intentionally minimal and orthogonal:
//
//   - InputBuilder      - assembles a *pkg.HookInput / *pkg.ApplicationHookInput.
//   - StaticSnapshots   - in-memory pkg.Snapshots backed by JSON literals.
//   - JQRun             - apply a JQ filter to JSON or to a Go value.
//   - PreparePatchCollector - construct a patch collector mock with sane defaults.
//   - PatchOperations   - decode the operations a hook recorded on a PatchCollector.
//
// All helpers play nicely with both *testing.T and Ginkgo's GinkgoT().
package helpers
