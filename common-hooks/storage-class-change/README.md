# `storage-class-change`

Module hook that watches PVCs, Pods, and StorageClasses for a single workload and:

1. Computes the **effective** storage class for the workload, layering cluster default → global config → module config → in-cluster PVC.
2. Stores the result at a well-known internal values path so Helm templates can branch on it.
3. **Evicts** pods whose PVCs have been deleted out of band.
4. **Deletes** the workload's PVCs and the workload itself (StatefulSet / Deployment / Prometheus) when the storage class actually changed, so the controller recreates everything with the new class.
5. Exports a `d8_emptydir_usage` Prometheus metric that is `1` when the module falls back to `emptyDir` (no storage class).

## What it watches

| Snapshot | Kind | Filter |
| --- | --- | --- |
| `pvcs` | `PersistentVolumeClaim` in `args.Namespace` | label `args.LabelSelectorKey=args.LabelSelectorValue` |
| `pods` | `Pod` in `args.Namespace` | same label selector |
| `storageClasses` | `StorageClass` (cluster-scoped) | none |

Each snapshot is JQ-filtered into a small struct with just the fields the hook needs (`name`, `namespace`, `storageClassName`, `isDeleted`, …).

## Effective storage class lookup

For a hook initialised with `args`, the effective storage class is computed in this order (later wins):

1. The cluster's **default** StorageClass (annotation `storageclass.kubernetes.io/is-default-class=true` or its beta variant).
2. `global.modules.storageClass` from **config values**, if set.
3. The class **currently bound to the workload's PVCs** (if any).
4. `<args.ModuleName>.<args.D8ConfigStorageClassParamName | "storageClass">` from **config values**, if set.

The result is written to `<lowerCamel(args.ModuleName)>.internal.<args.InternalValuesSubPath?>.effectiveStorageClass`. If the effective class is empty or the literal string `"false"`, the hook stores the boolean `false` instead and bumps the `d8_emptydir_usage` metric.

## Side effects

When the effective class differs from the currently-bound one:

- For each existing PVC the hook **deletes** the PVC (so the controller recreates it).
- The hook **deletes** the workload itself, dispatched on `args.ObjectKind`:

| `args.ObjectKind` | Effect |
| --- | --- |
| `StatefulSet` | `appsv1.StatefulSet` deleted via the K8s client. |
| `Deployment` | `appsv1.Deployment` deleted via the K8s client. |
| `Prometheus` | `monitoring.coreos.com/v1/prometheuses` deleted via the dynamic client. |
| anything else | The hook returns `unknown object kind <Kind>`. |

When a PVC has a `deletionTimestamp` but a Pod still references it, the hook issues a `policy/v1 Eviction` against the Pod.

## Usage

```go
package hooks

import (
    sccc "github.com/deckhouse/module-sdk/common-hooks/storage-class-change"
)

var _ = sccc.RegisterHook(sccc.Args{
    ModuleName:         "myModule",
    Namespace:          "d8-my-module",
    LabelSelectorKey:   "app",
    LabelSelectorValue: "data",
    ObjectKind:         "StatefulSet",
    ObjectName:         "data-set",

    // Optional knobs
    InternalValuesSubPath:         "data",            // → myModule.internal.data.effectiveStorageClass
    D8ConfigStorageClassParamName: "dataStorageClass", // → myModule.dataStorageClass instead of .storageClass
    BeforeHookCheck: func(input *pkg.HookInput) bool {
        // Skip the hook entirely when, e.g., the module is disabled.
        return input.Values.Get("myModule.enabled").Bool()
    },
})
```

## Hook configuration

- **Trigger:** `OnBeforeHelm` with `Order: 1`.
- **Snapshots:** `pvcs`, `pods`, `storageClasses` as described above.

## Testing

This hook has both unit and functional coverage:

- [`hook_test.go`](./hook_test.go) — table-driven JQ-filter tests using [`testing/helpers`](../../testing/helpers).
- [`hook_framework_test.go`](./hook_framework_test.go) — end-to-end scenarios driven through [`testing/framework`](../../testing/framework):
  - default StorageClass writes the right effective value;
  - explicit `global.modules.storageClass` overrides the default;
  - label selector scopes which PVCs participate;
  - `BeforeHookCheck` short-circuits the hook.
