# `testing/helpers` — small building blocks for hook unit tests

`testing/helpers` is the **unit-test layer** of the Module SDK testing toolkit. Where [`testing/framework`](../framework) runs hooks against a fake Kubernetes cluster, helpers stay much closer to the metal:

- `InputBuilder` — assembles a `*pkg.HookInput` with sensible defaults.
- `StaticSnapshots` — in-memory `pkg.Snapshots` backed by JSON / YAML / Go values.
- `RecordingPatchCollector` — `pkg.PatchCollector` that records every call for later inspection.
- `NewValues*` — real `pkg.PatchableValuesCollector` seeded from a JSON / YAML / map.
- `JQRunOnString` / `JQRunOnObject` — apply a JQ filter and decode the result in one call.

These helpers are deliberately small and orthogonal — pick the ones you need and ignore the rest.

## When to use it

Use helpers for **unit tests** that focus on a single hook handler:

- you know exactly which snapshots / values / patches the hook should see;
- you want to run a hook in microseconds without touching the fake K8s cluster;
- you are testing a JQ filter or a small piece of hook logic in isolation.

For **functional tests** that drive the whole pipeline (cluster YAML → snapshots → hook → cluster mutations), reach for [`testing/framework`](../framework) instead.

## Quick start

```go
package myhook_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/deckhouse/module-sdk/testing/helpers"
    myhook "example.com/mymodule"
)

func TestMyHook(t *testing.T) {
    in := helpers.NewInputBuilder(t).
        WithSnapshot("nodes",
            helpers.SnapshotJSON(`{"name":"node-a"}`),
            helpers.SnapshotJSON(`{"name":"node-b"}`),
        ).
        WithValuesJSON(`{"my":{"existing":"value"}}`).
        WithConfigValuesJSON(`{"module":{"enabled":true}}`).
        WithRecordingPatchCollector().
        WithCapturedLogger().
        Build()

    require.NoError(t, myhook.Handler(context.Background(), in))

    // Values
    assert.Equal(t, "value", in.Values.Get("my.existing").String())
    require.Len(t, in.Values.GetPatches(), 1)

    // PatchCollector
    pc := /* the same builder */ .RecordingPatchCollector()
    require.Len(t, pc.Recorded(), 2)
    assert.Equal(t, "Create", pc.Recorded()[0].Op)

    // Logs
    assert.Contains(t, /* builder */ .LogBuffer().String(), "expected log line")
}
```

## API at a glance

### `InputBuilder`

```go
b := helpers.NewInputBuilder(t)

b.WithSnapshot("nodes", helpers.SnapshotJSON(`{...}`))   // append
b.WithSnapshots(helpers.NewSnapshots())                  // replace map
b.WithValuesJSON(`{}`)                                   // or YAML / map
b.WithConfigValuesJSON(`{}`)                             // or YAML
b.WithRecordingPatchCollector()                          // typed RecordingPatchCollector
b.WithPatchCollector(myMock)                             // any pkg.PatchCollector
b.WithMetricsCollector(myMock)                           // any pkg.MetricsCollector
b.WithDependencyContainer(myDC)                          // any pkg.DependencyContainer
b.WithLogger(myLogger)                                   // any pkg.Logger
b.WithCapturedLogger()                                   // *log.Logger writing into a buffer

in := b.Build()                                          // *pkg.HookInput

// Accessors after Build:
b.Snapshots()                  // StaticSnapshots
b.Values()                     // pkg.PatchableValuesCollector
b.ConfigValues()               // pkg.PatchableValuesCollector
b.RecordingPatchCollector()    // *RecordingPatchCollector or nil
b.LogBuffer()                  // *bytes.Buffer or nil
```

Defaults: empty snapshots, empty values + config values, a `RecordingPatchCollector`, `metric.NewCollector`, `log.NewNop()`.

### Snapshots

```go
helpers.NewSnapshots()                          // empty StaticSnapshots
    .Add("k", helpers.SnapshotJSON(`{...}`))    // append
    .Set("k", snaps...)                         // replace bucket

helpers.SnapshotJSON(`{"name":"x"}`)            // pkg.Snapshot from raw JSON
helpers.SnapshotYAML("name: x")                 // pkg.Snapshot from YAML
helpers.SnapshotFromObject(myStruct)            // pkg.Snapshot from a Go value
helpers.SnapshotFromObjects([]MyType{...})      // []pkg.Snapshot
```

`StaticSnapshots` implements `pkg.Snapshots`, so you can pass it directly into a `*pkg.HookInput` if you don't want the builder.

### Values

```go
helpers.NewValues(map[string]any{...})                   // real PatchableValues, with map seed
helpers.NewValuesFromJSON(`{"foo":{"bar":"baz"}}`)       // same, from JSON
helpers.NewValuesFromYAML("foo:\n  bar: baz\n")          // same, from YAML

helpers.MarshalValues(v)                                 // JSON of v.GetPatches()
```

The store is a real `patchable-values.PatchableValues`, so:

- `v.Get(path)` returns a real `gjson.Result` from the seeded data;
- `v.Set(path, value)` records a real `add` patch op available via `v.GetPatches()`;
- `v.Remove(path)` records a real `remove` patch op when the path exists.

### `RecordingPatchCollector`

```go
pc := helpers.NewRecordingPatchCollector()

// hook calls pc.Create / pc.Delete / pc.PatchWith* …

pc.Recorded()                                  // []*RecordedOp in call order
pc.Filter("Delete", "DeleteInBackground")      // subset by op name
pc.Operations()                                // []pkg.PatchCollectorOperation
```

Each `RecordedOp` has the relevant fields populated for its op type:
- `Op` — `"Create"`, `"CreateOrUpdate"`, `"CreateIfNotExists"`, `"Delete"`, `"DeleteInBackground"`, `"DeleteNonCascading"`, `"JSONPatch"`, `"MergePatch"`, `"JQFilter"`.
- `Object` — the object passed to `Create*`.
- `APIVersion`, `Kind`, `Namespace`, `Name` — for `Delete*` and `Patch*`.
- `Patch`, `JQFilter`, `Options` — for the patch operations.

`RecordingPatchCollector` does **not** apply patches to anything — for that, use `testing/framework`.

### JQ helpers

```go
helpers.JQRunOnString(ctx, ".metadata.name", `{"metadata":{"name":"x"}}`, &out)
helpers.JQRunOnObject(ctx, ".spec.replicas", podSpec, &out)
```

Both compile the filter, run it against the input, and JSON-decode the result into `out` (which must be a non-nil pointer).

## Real-world examples in this repo

- [`testing/helpers/helpers_test.go`](./helpers_test.go) — exhaustive helper tests, doubles as documentation.
- [`common-hooks/copy-custom-certificate/hook_test.go`](../../common-hooks/copy-custom-certificate/hook_test.go), [`tls-certificate/order_certificate_test.go`](../../common-hooks/tls-certificate/order_certificate_test.go) — JQ filter tests.
- [`common-hooks/tls-certificate/internal_tls_test.go`](../../common-hooks/tls-certificate/internal_tls_test.go) — `InputBuilder` + real values store driving an entire certificate-rotation flow.
- [`examples/example-module/hooks/subfolder/patch_hook_test.go`](../../examples/example-module/hooks/subfolder/patch_hook_test.go) — `RecordingPatchCollector` asserting on op sequence.
- [`examples/example-module/hooks/subfolder/values_getting_hook_test.go`](../../examples/example-module/hooks/subfolder/values_getting_hook_test.go) — mix of helpers (happy path) and `mock.OutputPatchableValuesCollectorMock` (error paths).
- [`examples/dependency-example-module/hooks/subfolder/http_client_hook_test.go`](../../examples/dependency-example-module/hooks/subfolder/http_client_hook_test.go) — small `httpDC` factory wired through `InputBuilder.WithDependencyContainer`.
