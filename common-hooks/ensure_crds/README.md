# `ensure_crds`

Module hook that installs (or updates) the CustomResourceDefinitions shipped with your module on **every module startup**.

It is a thin wrapper around [`pkg/crd-installer`](../../pkg/crd-installer): the hook gets a Kubernetes client from the SDK's `DependencyContainer`, expands a glob to a list of CRD YAML files, and lets the installer apply them.

## What it does

On `OnStartup` (order `5`) the hook:

1. Calls `input.DC.GetK8sClient()` to obtain a real Kubernetes client.
2. Expands the configured CRD glob (e.g. `/deckhouse/modules/002-deckhouse/crds/*.yaml`) into a list of files.
3. Skips files whose basename starts with `doc-` (these are documentation-only manifests).
4. Applies all remaining CRDs via the installer, labelling them `heritage=deckhouse`.

If anything fails, the error is logged and bubbles up so addon-operator can surface it.

## Usage

```go
package hooks

import (
    ensure_crds "github.com/deckhouse/module-sdk/common-hooks/ensure_crds"
)

// Match all CRDs shipped with the module, except doc-*.yaml.
var _ = ensure_crds.RegisterEnsureCRDsHookEM(
    "/deckhouse/modules/002-mymodule/crds/*.yaml",
)
```

The exported public surface is small:

| Function | Purpose |
| --- | --- |
| `RegisterEnsureCRDsHookEM(crdsGlob string) bool` | Registers the hook in the SDK's global registry. **Recommended for external modules.** |
| `EnsureCRDsHandler(crdsGlob string) func(ctx, *pkg.HookInput) error` | Returns the bare handler without registering anything; useful for composing custom hooks. |
| `EnsureCRDs(ctx, input, crdsGlob)` | Lower-level entry point that expects a ready-built `*pkg.HookInput`. |

## File filter

The default filter excludes manifests whose basename starts with `doc-`:

```go
crdinstaller.WithFileFilter(func(crdFilePath string) bool {
    return !strings.HasPrefix(filepath.Base(crdFilePath), "doc-")
})
```

If you ship CRD YAMLs with extra non-installable files (release notes, examples, …), prefix them with `doc-`.

## Hook configuration

- **Trigger:** `OnStartup` with `Order: 5`.
- **No Kubernetes bindings** — the hook talks to the API server directly via `input.DC.GetK8sClient()`.
- **No values output** — the side effect is purely on the cluster.

## Testing

The CRD-installer logic itself is exercised by [`pkg/crd-installer/installer_test.go`](../../pkg/crd-installer/installer_test.go). Module-level tests should focus on supplying the right glob and rely on those installer tests for the rest.

For functional coverage, the hook can be driven through [`testing/framework`](../../testing/framework) — register CRDs from a temp directory and assert that the framework's fake cluster contains the resulting `CustomResourceDefinition` objects after `RunHook()`.
