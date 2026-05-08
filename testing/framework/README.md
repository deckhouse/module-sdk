# `testing/framework` — deckhouse-style hook test harness

Functional, end-to-end testing for module hooks **without** a real Kubernetes cluster, addon-operator, or shell-operator.

The framework spins up a fake dynamic Kubernetes client, generates snapshots from your hook's `KubernetesConfig` bindings, runs the handler with a real `*pkg.HookInput`, and replays every recorded patch operation against the fake cluster — so you can assert on cluster state directly.

It mirrors the flow of [`deckhouse/testing/hooks`](https://github.com/deckhouse/deckhouse/tree/main/testing/hooks), but has no dependency on `addon-operator` or `shell-operator`.

## When to use it

Use the framework for **functional tests** of a hook:

- the hook reads multiple kinds of resources via `KubernetesConfig` bindings;
- the hook produces a sequence of `Create` / `Delete` / `Patch` operations and you want to assert on the resulting cluster state;
- the hook uses `input.DC.GetK8sClient()` to interact with the API server directly;
- you want a single test to walk through several state transitions (`KubeStateSet` → `RunHook` → assert → `KubeStateSet` → `RunHook` → assert).

For small unit tests where you only care about a single path, prefer [`testing/helpers`](../helpers).

## Quick start

```go
package myhook_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/require"

    "github.com/deckhouse/module-sdk/pkg"
    "github.com/deckhouse/module-sdk/testing/framework"
)

func TestMyHook(t *testing.T) {
    cfg := &pkg.HookConfig{
        Kubernetes: []pkg.KubernetesConfig{{
            Name:       "nodes",
            APIVersion: "v1",
            Kind:       "Node",
            JqFilter:   `{name: .metadata.name}`,
        }},
    }

    handler := func(_ context.Context, in *pkg.HookInput) error {
        in.Values.Set("count", len(in.Snapshots.Get("nodes")))
        return nil
    }

    f := framework.HookExecutionConfigInit(t, cfg, handler, `{}`, `{}`)

    f.KubeStateSet(`
---
apiVersion: v1
kind: Node
metadata:
  name: kube-worker-1
---
apiVersion: v1
kind: Node
metadata:
  name: kube-worker-2
`)

    f.RunHook()

    require.NoError(t, f.HookError())
    require.Equal(t, int64(2), f.ValuesGet("count").Int())
}
```

## API reference

### Construction

| Function | Purpose |
| --- | --- |
| `HookExecutionConfigInit(t, cfg, handler, initValues, initConfigValues)` | Deckhouse-compatible constructor. `initValues` / `initConfigValues` accept JSON or YAML; pass `"{}"` if not needed. |
| `NewHookExecutionConfig(t, cfg, handler, opts...)` | Same, but with explicit `Option`s. Accepts `WithInitialValues`, `WithInitialConfigValues`, `WithSchemeBuilder`, `WithCRD`, `WithOpenAPIDir`, `WithValuesSchema`, `WithConfigValuesSchema`. |

`t` is a `testing.TB`, so `*testing.T`, sub-tests, and `GinkgoT()` all work.

### Cluster state

| Method | Purpose |
| --- | --- |
| `KubeStateSet(yaml)` | Replace **all** objects in the fake cluster with the resources defined in the multi-document YAML. Each `RunHook` after this regenerates snapshots from the new state. |
| `AddKubeObject(yaml)` | Add objects without resetting state. |
| `KubernetesResource(kind, namespace, name)` | Fetch a current object as `*unstructured.Unstructured` (or `nil` if not found). |
| `KubernetesGlobalResource(kind, name)` | Same, for cluster-scoped resources. |
| `KubeClient()` | Raw `dynamic.Interface` — escape hatch when you need to seed something the YAML loader cannot express. |

### Custom resources

```go
f := framework.NewHookExecutionConfig(t, cfg, handler,
    framework.WithSchemeBuilder(myapis.SchemeBuilder), // for typed CRDs
    framework.WithCRD("acme.io", "v1", "Widget", true), // for ad-hoc CRs
)

// or, after construction:
f.RegisterCRD("acme.io", "v1", "Widget", true)
```

`WithSchemeBuilder` is preferred when the CRD has Go types you can import; `WithCRD` / `RegisterCRD` is used to teach the GVR resolver about a kind that lives only in YAML.

### OpenAPI defaults

In production, addon-operator applies defaults from the module's OpenAPI schemas (`openapi/values.yaml` and `openapi/config-values.yaml`) before invoking a hook. The framework can do the same so tests don't drift from real-world behaviour:

```go
f := framework.NewHookExecutionConfig(t, cfg, handler,
    framework.WithOpenAPIDir("../openapi"),
    framework.WithInitialValues(`{"https": {"mode": "CertManager"}}`),
)
```

Behaviour:

- `WithOpenAPIDir(dir)` looks for `<dir>/values.yaml` and `<dir>/config-values.yaml`. Whichever ones are present are loaded.
- For each schema, the framework extracts every `default:` declared in it and uses the result as a baseline values document.
- Anything passed via `WithInitialValues` / `WithInitialConfigValues` is then deep-merged on top — your test's values always override schema defaults.
- The `x-extend` extension is honoured. If `values.yaml` declares `x-extend.schema: config-values.yaml`, the values store inherits all defaults from `config-values.yaml` plus its own.

For more granular control use `WithValuesSchema(path)` / `WithConfigValuesSchema(path)` instead — they fail the test if the file is missing.

The lower-level helpers `LoadOpenAPISchema`, `SchemaDefaults`, and `MergeValues` are also exported, which is handy when you want to assemble a full values document outside `NewHookExecutionConfig`:

```go
schema, err := framework.LoadOpenAPISchema("../openapi/values.yaml")
require.NoError(t, err)
defaults := framework.SchemaDefaults(schema)
merged := framework.MergeValues(defaults, map[string]any{
    "replicas": 5,
})
```

### Values

| Method | Purpose |
| --- | --- |
| `ValuesGet(path) gjson.Result` | Read current values at a dotted path. |
| `ConfigValuesGet(path) gjson.Result` | Same, for `ConfigValues`. |
| `ValuesSet(path, any)` / `ConfigValuesSet(path, any)` | Set a value (persists across `RunHook` calls). |
| `ValuesSetFromYaml(path, []byte)` / `ConfigValuesSetFromYaml(path, []byte)` | Same, but parses YAML. |
| `ValuesDelete(path)` / `ConfigValuesDelete(path)` | Remove a path. |
| `ValuesJSON()` / `ConfigValuesJSON()` | Whole-document JSON for snapshot-style assertions. |

### Running and inspecting

| Method | Purpose |
| --- | --- |
| `RunHook()` / `RunHookCtx(ctx)` | Generate snapshots, build `HookInput`, invoke the handler, apply values patches, replay cluster patches. |
| `HookError() error` | Error returned by the handler from the most recent `RunHook`. |
| `Snapshots() pkg.Snapshots` | Snapshots that were passed to the hook. |
| `PatchedOperations() []RecordedPatch` | Typed view of every `Create`/`Delete`/`Patch` issued by the hook. |
| `PatchOperations() []pkg.PatchCollectorOperation` | The same, but cast to the `pkg.PatchCollectorOperation` interface. |
| `CollectedMetrics() []MetricOperation` | Metric operations emitted via `input.MetricsCollector`. |
| `Logger() *log.Logger` / `LoggerOutput() *bytes.Buffer` | Test logger and its captured output. |
| `DependencyContainer()` | The framework's DC. Use `SetHTTPClient`, `SetRegistryClient`, `SetClock` to inject mocks before `RunHook`. |

## How it works

`RunHook` runs the same five-step pipeline every time:

1. **Generate snapshots.** For each `KubernetesConfig` binding, the framework lists matching resources from the fake cluster (honouring `NameSelector`, `NamespaceSelector`, `LabelSelector`, `FieldSelector`), then runs `JqFilter` on each match.
2. **Build a real `HookInput`.** Values and config values are wrapped in [`pkg/patchable-values.PatchableValues`](../../pkg/patchable-values), the patch collector is a `recordingPatchCollector`, and the metrics collector is a real `internal/metric.Collector`.
3. **Invoke the handler.** Errors are captured in `HookError()`.
4. **Apply values patches.** The framework merges the patches the hook produced via `input.Values.Set/Remove` back into its values store (and same for config values).
5. **Replay cluster patches.** Each recorded `Create` / `Delete` / `MergePatch` / `JSONPatch` / `JQFilter` is applied to the fake dynamic client, so `KubernetesResource(...)` returns the post-hook state.

If the handler returned an error, step 5 is skipped — error-path tests can still assert on values patches and the recorded operations the hook *intended* to issue.

## Pitfalls and tips

- The fake client uses `meta.UnsafeGuessKindToResource` for GVR mapping. Standard Kubernetes kinds (`Pod`, `Node`, `StatefulSet`, …) work out of the box; custom kinds need `WithCRD` or `WithSchemeBuilder`.
- `KubeStateSet` rebuilds the fake client; if you keep references to objects fetched before, refresh them with `KubernetesResource`.
- The `DependencyContainer`'s HTTP and registry clients return errors by default. If your hook calls `input.DC.GetHTTPClient()` you must override them via `f.DependencyContainer().SetHTTPClient(...)` before `RunHook`.
- `LoggerOutput()` captures everything the hook logs, including the framework's own diagnostic messages — use `strings.Contains` rather than line-by-line equality.

## Real-world examples in this repo

- [`testing/framework/example_test.go`](./example_test.go) — verbose, deliberately documentation-style end-to-end test.
- [`common-hooks/storage-class-change/hook_framework_test.go`](../../common-hooks/storage-class-change/hook_framework_test.go) — selectors, config-values overrides, `BeforeHookCheck` gating.
- [`examples/example-module/hooks/subfolder/snapshot_framework_test.go`](../../examples/example-module/hooks/subfolder/snapshot_framework_test.go) — snapshot binding driven by cluster YAML.
- [`examples/example-module/hooks/subfolder/patch_framework_test.go`](../../examples/example-module/hooks/subfolder/patch_framework_test.go) — replay of `Create`+`Delete`+`Patch` operations.
- [`examples/single-file-example/hooks/main_framework_test.go`](../../examples/single-file-example/hooks/main_framework_test.go) — labels and namespace scoping.
- [`examples/dependency-example-module/hooks/subfolder/http_client_framework_test.go`](../../examples/dependency-example-module/hooks/subfolder/http_client_framework_test.go) — overriding the HTTP client on the framework's `DependencyContainer`.
