# `testing/` — Tools for testing module hooks

This directory contains everything you need to test hooks built on top of the Module SDK without spinning up a real Kubernetes cluster, addon-operator, or shell-operator.

It is organised around **three concentric layers**, picked depending on how much of the hook pipeline you want to exercise:

| Package | Use it when… | Style |
| --- | --- | --- |
| [`testing/mock`](./mock) | You only need to substitute one or two collaborators (`PatchCollector`, `Snapshots`, `DependencyContainer`, `KubernetesClient`, …) | minimock-generated mocks |
| [`testing/helpers`](./helpers) | You want a small, focused unit test for a single hook handler with realistic values, snapshots, patches, and logger | Builder + ready-made fakes |
| [`testing/framework`](./framework) | You want a deckhouse-style end-to-end test: declare cluster YAML, run the hook, assert on snapshots / values / patched cluster state | Full pipeline with a fake K8s cluster |

A fuller description of when to pick each layer is in the project-level [`TESTING.md`](../TESTING.md).

## Quick orientation

```text
testing/
├── mock/        # auto-generated mocks for every pkg.* interface
├── helpers/     # small, hand-written building blocks for unit tests
└── framework/   # deckhouse-style harness: fake k8s cluster + hook runner
```

- **`mock`** is generated with [`minimock`](https://github.com/gojuno/minimock) from the interfaces in [`../pkg`](../pkg). It is the lowest-level layer — you compose individual mocks yourself and assemble a `*pkg.HookInput`.
- **`helpers`** provides `InputBuilder`, `StaticSnapshots`, `RecordingPatchCollector`, `NewValues*`, and `JQRunOn*`. These are thin layers on top of the real implementations from `pkg/*` so the values store actually records patches, the snapshot really decodes JSON, etc.
- **`framework`** is the heaviest of the three. It runs the hook end-to-end: it owns a fake dynamic Kubernetes client, generates snapshots from the hook's `KubernetesConfig` bindings, replays the patches the hook recorded, and lets you assert on the resulting cluster state.

## Example matrix

The same hook tested at each layer:

```go
// 1) testing/mock — full control, full ceremony
snapshots := mock.NewSnapshotsMock(t).GetMock.When("nodes").Then(...)
values    := mock.NewOutputPatchableValuesCollectorMock(t)
input     := &pkg.HookInput{Snapshots: snapshots, Values: values, Logger: log.NewNop()}
require.NoError(t, MyHook(ctx, input))

// 2) testing/helpers — same intent, less ceremony, real values store
input := helpers.NewInputBuilder(t).
    WithSnapshot("nodes", helpers.SnapshotJSON(`{"name":"n1"}`)).
    WithValuesJSON(`{}`).
    Build()
require.NoError(t, MyHook(ctx, input))
require.Equal(t, "n1", input.Values.Get("found.node").String())

// 3) testing/framework — drive from cluster YAML, replay patches
hec := framework.HookExecutionConfigInit(t, cfg, MyHook, `{}`, `{}`)
hec.KubeStateSet(`
apiVersion: v1
kind: Node
metadata: {name: n1}
`)
hec.RunHook()
require.NoError(t, hec.HookError())
require.Len(t, hec.Snapshots().Get("nodes"), 1)
```

## Choosing a layer

- Need to assert that **a specific JSON path was set** on values? → `helpers.NewValuesFromJSON` (unit test).
- Need to assert that the hook produced the **right sequence of patch operations** without applying them? → `helpers.RecordingPatchCollector` (unit test).
- Need to assert that **a real cluster ends up with the expected state** after Create+Delete+Patch? → `framework` (functional test).
- Need to inject **a precise sequence of mock return values** for one specific dependency? → `mock` (low-level test).

Most modules end up using a mix of `helpers` for fast unit tests and a handful of `framework` tests for the trickiest end-to-end paths.

## See also

- [`TESTING.md`](../TESTING.md) — project-wide testing philosophy.
- [`testing/framework/README.md`](./framework/README.md) — deckhouse-style harness reference.
- [`testing/helpers/README.md`](./helpers/README.md) — unit-test helper reference.
- [`pkg/jq`](../pkg/jq) — the JQ engine used by snapshot filters; useful when testing JQ expressions in isolation.
