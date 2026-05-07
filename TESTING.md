# Testing Module SDK hooks

This document describes how we test hooks built on top of the Module SDK.

The goal is simple: **let module developers test hook logic with the same speed and confidence as plain Go code**, without standing up a real Kubernetes cluster, addon-operator, or shell-operator.

## TL;DR

There are three layers, picked by the size of the test you want to write:

| Layer | Use it for | Speed | Lives at |
| --- | --- | --- | --- |
| **Mocks** | Tests that need precise control over a single dependency | µs | [`testing/mock`](./testing/mock) |
| **Helpers** | Unit tests of a single hook handler | µs | [`testing/helpers`](./testing/helpers) |
| **Framework** | Functional tests driving the whole hook pipeline against a fake K8s cluster | ms | [`testing/framework`](./testing/framework) |

Pick the smallest layer that lets you write the assertion you actually care about.

## Test pyramid

```text
                ┌──────────────────────┐
                │  framework  (slow)   │   end-to-end behaviour
                ├──────────────────────┤
                │  helpers   (fast)    │   handler-level units
                ├──────────────────────┤
                │  mock     (fastest)  │   single-collaborator
                └──────────────────────┘
```

Most modules end up with a wide base of `helpers`-based unit tests and a thin layer of `framework` functional tests for the trickiest paths.

## Layer 1 — `testing/mock`

`testing/mock` is generated with [`minimock`](https://github.com/gojuno/minimock) from the interfaces in [`pkg`](./pkg). It is the lowest-level layer: you compose mocks yourself and assemble a `*pkg.HookInput` by hand.

Use it when you need **precise control** over a single dependency:

```go
values := mock.NewOutputPatchableValuesCollectorMock(t)
values.GetMock.When("global.discovery.clusterDomain").
    Then(gjson.Result{Type: gjson.String, Str: "cluster.local"})

input := &pkg.HookInput{
    Values: values,
    Logger: log.NewNop(),
}
require.NoError(t, MyHook(ctx, input))
```

This layer is the right call when:

- you want to assert on the **exact sequence** of calls a hook makes against a collaborator;
- you need to inject an **error from a specific call** (e.g. `ArrayCount → error`);
- the hook is small enough that a real values store is overkill.

## Layer 2 — `testing/helpers`

`testing/helpers` is the **default** unit-test layer. It bundles a set of small, hand-written building blocks on top of the real implementations from `pkg/*`:

- `InputBuilder` — fluent assembly of `*pkg.HookInput`.
- `StaticSnapshots` — in-memory `pkg.Snapshots` backed by JSON / YAML / Go values.
- `RecordingPatchCollector` — a `pkg.PatchCollector` that records every call.
- `NewValuesFromJSON/YAML/Map` — a real `pkg.PatchableValuesCollector` seeded from your input.
- `JQRunOnString/Object` — apply a JQ filter and decode the result in one call.

Typical test:

```go
in := helpers.NewInputBuilder(t).
    WithSnapshot("nodes", helpers.SnapshotJSON(`{"name":"n1"}`)).
    WithValuesJSON(`{"my":{"existing":"value"}}`).
    WithRecordingPatchCollector().
    Build()

require.NoError(t, MyHook(context.Background(), in))

// Real values store: actual patches were recorded.
require.Len(t, in.Values.GetPatches(), 1)
```

Use this layer when:

- you know exactly what the hook should see and do;
- you want to test the **happy path and a few error paths** of a single handler;
- you're writing **a JQ filter test** with `helpers.JQRunOn*`.

A complete reference is at [`testing/helpers/README.md`](./testing/helpers/README.md).

## Layer 3 — `testing/framework`

`testing/framework` is the **functional-test** layer, mirroring [`deckhouse/testing/hooks`](https://github.com/deckhouse/deckhouse/tree/main/testing/hooks). The framework:

1. owns a **fake dynamic Kubernetes client** seeded from YAML;
2. **generates snapshots** from the hook's `KubernetesConfig` bindings (selectors + JQ);
3. runs the handler with a real `*pkg.HookInput`;
4. **applies values patches** back to the values store;
5. **replays cluster patches** (`Create` / `Delete` / `Patch`) against the fake cluster.

After `RunHook`, you assert on:
- snapshots passed in,
- final values & config values,
- recorded patch operations,
- post-hook cluster state via `KubernetesResource(...)`,
- collected metrics,
- captured logs.

Typical test:

```go
f := framework.HookExecutionConfigInit(t, cfg, handler, `{}`, `{}`)
f.KubeStateSet(`
---
apiVersion: v1
kind: Pod
metadata: {name: app, namespace: default}
status: {phase: Running}
`)
f.RunHook()

require.NoError(t, f.HookError())
require.NotNil(t, f.KubernetesResource("ConfigMap", "default", "app-status"))
```

Use this layer when:

- the hook reads multiple bindings or relies on label / namespace selectors;
- the hook ends up issuing several patch operations and you want to verify the **resulting cluster state**;
- the hook talks to the API server through `input.DC.GetK8sClient()`;
- you want one test to walk through several state transitions.

A complete reference is at [`testing/framework/README.md`](./testing/framework/README.md).

## Picking a layer in practice

A small flowchart:

```text
Are you only testing a JQ filter?
  └─► helpers.JQRunOn{String,Object}

Are you testing a single handler in isolation,
with snapshots and values you can describe inline?
  └─► helpers.NewInputBuilder + RecordingPatchCollector

Do you need to assert on the resulting Kubernetes objects
(Create + Delete chains, Patch results, JQ mutations)?
  └─► testing/framework

Do you need to inject a very specific failure from one
collaborator (e.g. ArrayCount returning an error)?
  └─► testing/mock + a hand-built *pkg.HookInput
```

It's normal — and expected — to mix layers in the same package. See e.g. [`examples/example-module/hooks/subfolder`](./examples/example-module/hooks/subfolder), where:

- `*_test.go` files use `helpers.NewInputBuilder` for the bulk of unit tests;
- `*_framework_test.go` files use `testing/framework` for end-to-end coverage;
- a handful of error-path tests fall back to `testing/mock` for a specific failure.

## Project-wide testing conventions

- **No global state.** Tests should not rely on `registry.Registry()` having a particular content; build a `*pkg.HookConfig` locally in the test or share it via a small helper in the package.
- **Ginkgo is being phased out.** New tests should use plain `*testing.T` + `testify`. Existing Ginkgo suites are migrated when their package is touched.
- **Real values stores beat value mocks.** If you can use `helpers.NewValuesFromJSON`, do — the assertions become "did the hook produce the right `add`/`remove` operations?", which is more honest than "did the hook call `Set` with these arguments?".
- **JQ filters are tested in isolation** via `helpers.JQRunOn{String,Object}`. This keeps the filter expression visible in the test source and decouples it from the hook handler.
- **Functional tests are sparse but high-value.** Aim for a handful of `framework` tests per hook, not one per code path.
- **Lint and vet are required.** `go vet ./...` and `golangci-lint run ./...` must stay green; this is enforced by `make test` and `make lint`.

## Running the tests

```sh
# Module SDK and common-hooks
make test
make lint

# Each example module is a standalone Go module under examples/.
make examples
```

Each example module has its own `go.mod` and its own test suite — they are deliberately self-contained so they double as documentation.

## See also

- [`testing/README.md`](./testing/README.md) — overview of the testing tree.
- [`testing/framework/README.md`](./testing/framework/README.md) — functional-test harness reference.
- [`testing/helpers/README.md`](./testing/helpers/README.md) — unit-test helper reference.
- [`pkg/jq`](./pkg/jq) — JQ engine, useful when you want to debug a snapshot filter expression.
