# `copy-custom-certificate`

Module hook that copies a user-supplied TLS certificate from a Secret in `d8-system` into the module's internal values, **only** when the module is configured to use the `CustomCertificate` HTTPS mode.

## What it does

1. Watches Secrets in the `d8-system` namespace, ignoring the helm-owned ones (selector `owner notin (helm)`).
2. Filters each matched Secret with the JQ expression `JQFilterCustomCertificate`, which extracts `name`, `key`, `crt`, and `ca` from `.data`.
3. On every run:
    - If no certificates are seen, the hook logs and exits.
    - If the module's effective HTTPS mode is **not** `CustomCertificate`, it removes the previously-set internal value.
    - If the configured `secretName` matches one of the discovered Secrets, the hook writes the cert payload to `<moduleName>.internal.customCertificateData`.
    - If the configured `secretName` is set but no Secret with that name exists, the hook returns an error.

## Resulting values

The hook writes the certificate at `<moduleName>.internal.customCertificateData`:

```yaml
<moduleName>:
  internal:
    customCertificateData:
      ca.crt: |
        ...
      tls.crt: |
        ...
      tls.key: |
        ...
```

## Configuration paths the hook reads

| Path | Meaning |
| --- | --- |
| `<moduleName>.https.customCertificate.secretName` (config) | Module-level override of the secret name. |
| `global.modules.https.customCertificate.secretName` (config) | Cluster-wide fallback. |
| `<moduleName>` HTTPS mode | Computed by `pkg/utils/patchable-values.GetHTTPSMode(input, moduleName)`. |

The hook uses `GetValuesFirstDefined`, so the module-level path wins over the global one.

## Usage

```go
package hooks

import (
    cc "github.com/deckhouse/module-sdk/common-hooks/copy-custom-certificate"
)

// Registers a hook that copies CustomCertificate Secrets into
// "myModule.internal.customCertificateData" when HTTPS mode is enabled.
var _ = cc.RegisterHook("myModule")
```

## Hook configuration

The hook registers with:

- **Order:** `OnBeforeHelm.Order = 10`
- **Bindings:** Secrets in `d8-system` whose `owner` label is anything but `helm`.

## JQ filter

```jq
{
  "name": .metadata.name,
  "key":  .data."tls.key",
  "crt":  .data."tls.crt",
  "ca":   .data."ca.crt"
}
```

The filter is exported as `JQFilterCustomCertificate` so other hooks (or tests) can reuse it.

## Testing

Tests live in [`hook_test.go`](./hook_test.go) and use the JQ helper from [`testing/helpers`](../../testing/helpers) to validate the filter against a canonical `kubernetes.io/tls` Secret.
