# `common-hooks/` — reusable Module SDK hooks

`common-hooks/` ships a curated set of hooks that solve recurring needs across Deckhouse modules. Importing one of them is usually a one-liner: each package exports a `RegisterHook(...)` (or similar) function that takes module-specific arguments and registers a fully-configured hook with the SDK's global registry.

## Available hooks

| Package | What it does |
| --- | --- |
| [`copy-custom-certificate`](./copy-custom-certificate) | Copies user-provided TLS certificates from `d8-system` Secrets into the module's internal values, gated by the module's HTTPS mode. |
| [`ensure_crds`](./ensure_crds) | Installs (or updates) all CRD YAMLs matched by a glob, on module startup. |
| [`external_auth`](./external_auth) | Wires up Dex-based or user-provided external authentication settings into the module values. |
| [`storage-class-change`](./storage-class-change) | Tracks storage-class changes, evicts pods on stale PVCs, and re-creates StatefulSets / Deployments / Prometheuses when the effective storage class changes. |
| [`tls-certificate`](./tls-certificate) | Generates or refreshes self-signed TLS certificates and orders Kubernetes-signed certificates via the `certificates.k8s.io` API. |

## Usage shape

Most common hooks follow the same shape:

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
})
```

A few notes:

- The `_ = ...RegisterHook(...)` idiom registers the hook at package init via the SDK registry. Importing the package is enough to enable the hook.
- Each hook embeds its own `pkg.HookConfig` (binding contexts, schedules, JQ filters, …); you only supply module-level parameters.
- For application hooks use the `…Application` variants where they exist; module hooks use the `pkg.HookInput` shape.

## Testing

Each common hook ships with both **unit** and (where applicable) **functional** tests:

- Unit tests live next to the hook (`*_test.go`) and rely on [`testing/helpers`](../testing/helpers).
- Functional tests live in `*_framework_test.go` and use [`testing/framework`](../testing/framework) to drive the hook against a fake Kubernetes cluster.

See [`TESTING.md`](../TESTING.md) for the bigger picture.

## Examples

Working examples that consume these common hooks live under [`examples/common-hooks`](../examples/common-hooks).
