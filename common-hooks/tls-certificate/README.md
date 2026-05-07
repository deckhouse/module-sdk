# `tls-certificate`

Two complementary hooks for managing TLS material inside a Deckhouse module:

| Hook | Use it when… |
| --- | --- |
| **Self-signed (`internal_tls.go`)** | The module needs an in-cluster TLS pair for its own services (typically a webhook). The module signs the cert itself. |
| **Order from cluster CA (`order_certificate.go`)** | The module needs a certificate signed by the Kubernetes cluster CA via the `certificates.k8s.io` API. |

Both hooks store the resulting `ca.crt` / `tls.crt` / `tls.key` in module values, so Helm templates can render them into Secrets directly.

---

## Self-signed TLS — `RegisterInternalTLSHookEM`

Run on `OnBeforeHelm` (order `5`) and on a daily schedule (`42 4 * * *`).

The hook reads a TLS-typed Secret from the cluster (named `TLSSecretName` in `Namespace`) and:

- if **the Secret is missing**, generates a fresh CA + cert and writes them to values;
- if **the Secret is present and still valid**, copies it into values verbatim;
- if **the certificate or its CA is outdated** (within `CAOutdatedDuration` / `CertOutdatedDuration` of expiry), regenerates the matching half (CA only, cert only, or both);
- if **the certificate's CA expiry no longer matches the configured `CAExpiryDuration`**, regenerates everything;
- supports a **shared CA** via `CommonCAValuesPath` so multiple webhooks can sign with the same CA (with custom field names via `CommonCACertField` / `CommonCAKeyField`).

### Quick start

```go
package hooks

import (
    tlscert "github.com/deckhouse/module-sdk/common-hooks/tls-certificate"
)

var _ = tlscert.RegisterInternalTLSHookEM(tlscert.GenSelfSignedTLSHookConf{
    CN:            "example-webhook",
    TLSSecretName: "secret-webhook-cert",
    Namespace:     "d8-example-module",

    SANs: tlscert.DefaultSANs([]string{
        "example-webhook",
        "example-webhook.d8-example-module",
        "example-webhook.d8-example-module.svc",
        "%CLUSTER_DOMAIN%://example-webhook.d8-example-module.svc",
        "%PUBLIC_DOMAIN%://example-webhook.d8-example-module.svc",
    }),

    FullValuesPathPrefix: "exampleModule.internal.webhookCert",
})
```

### Resulting values

The hook writes to `FullValuesPathPrefix`:

```yaml
exampleModule:
  internal:
    webhookCert:
      ca:  ...PEM...
      crt: ...PEM...
      key: ...PEM...
```

In Helm templates, run these through `b64enc` before putting them into a Secret.

### `GenSelfSignedTLSHookConf` — the most useful fields

| Field | Purpose |
| --- | --- |
| `CN` | Common Name of the certificate; usually the module name. |
| `Namespace` / `TLSSecretName` | Where the persisted Secret lives. |
| `SANs` | Function returning the SAN list. Use `DefaultSANs([]string{...})` for `%CLUSTER_DOMAIN%` / `%PUBLIC_DOMAIN%` substitution. |
| `KeyAlgorithm` / `KeySize` | Defaults to `ecdsa` / `256`. |
| `Usages` | Defaults to `signing`, `key encipherment`, `requestheader-client`. |
| `FullValuesPathPrefix` | Where to store CA / Crt / Key in values. |
| `CommonCAValuesPath` | If set, share an existing CA stored at this path. |
| `CommonCACertField` / `CommonCAKeyField` | Field names within the shared-CA object (default `crt` / `key`). |
| `CommonCACanonicalName` | CN used when generating the shared CA, falls back to `CN`. |
| `CAExpiryDuration` / `CAOutdatedDuration` | Default 10y / 6mo. |
| `CertExpiryDuration` / `CertOutdatedDuration` | Default 10y / 6mo. |
| `BeforeHookCheck` | Optional gate; the hook is a no-op when this returns `false`. |

---

## Order a certificate from the cluster CA — `RegisterOrderCertificateHookEM`

Hooks for modules whose components must present certificates **signed by the cluster CA**, going through the `certificates.k8s.io/v1` `CertificateSigningRequest` (CSR) API.

The hook is registered with one or more `OrderCertificateRequest` entries. For each request it:

1. Reads the existing TLS Secret (snapshot `certificateSecrets`).
2. Compares the SAN list of the existing cert with the desired one.
3. If the cert is missing or its SANs no longer match, posts a CSR, **waits for it to be approved**, and stores the resulting `tls.crt` / `tls.key` (and `ca.crt`) in module values.

### Quick start

```go
package hooks

import (
    certificatesv1 "k8s.io/api/certificates/v1"

    tlscert "github.com/deckhouse/module-sdk/common-hooks/tls-certificate"
)

var _ = tlscert.RegisterOrderCertificateHookEM([]tlscert.OrderCertificateRequest{
    {
        Namespace:  "d8-example-module",
        SecretName: "myservice-client-tls",
        CommonName: "system:myservice",
        SignerName: "kubernetes.io/kube-apiserver-client",
        Groups:     []string{"system:masters"},
        Usages: []certificatesv1.KeyUsage{
            certificatesv1.UsageDigitalSignature,
            certificatesv1.UsageKeyEncipherment,
            certificatesv1.UsageClientAuth,
        },

        SANs: []string{
            "myservice.d8-example-module",
            "%CLUSTER_DOMAIN%://myservice.d8-example-module",
        },

        ModuleName: "exampleModule",
        ValueName:  "internal.myService.cert",
    },
})
```

### `OrderCertificateRequest`

| Field | Purpose |
| --- | --- |
| `Namespace` / `SecretName` | Existing Secret to inspect (the one your chart will render). |
| `CommonName`, `Groups`, `SANs`, `Usages`, `SignerName` | CSR contents. `%CLUSTER_DOMAIN%` and `%PUBLIC_DOMAIN%` placeholders in `SANs` are expanded from `global.discovery.clusterDomain` / `global.modules.publicDomainTemplate`. |
| `ModuleName` / `ValueName` | Where to write the resulting cert in values: `<ModuleName>.<ValueName>` ⇒ `{ ca, crt, key }`. |
| `WaitTimeout` | How long to wait for CSR approval (default `1m`, overridable per request). |
| `ExpirationSeconds` | Optional. Pass through to the CSR. |

### Hook configuration

- **Trigger:** `OnBeforeHelm` (order `5`) + daily schedule `42 4 * * *`.
- **Snapshots:** `certificateSecrets` — Secrets in the listed namespaces with the listed names, JQ-filtered with `JQFilterApplyCertificateSecret` to extract `name`, `key`, `crt`.

---

## JQ filters

This package exports two JQ filters that are useful in tests:

- `JQFilterTLS` — `{ key: .data."tls.key", crt: .data."tls.crt", ca: .data."ca.crt" }`
- `JQFilterApplyCertificateSecret` — like the above but accepts both `tls.*` and `client.*` keys, and includes the secret name.

## Testing

- [`internal_tls_test.go`](./internal_tls_test.go) — covers every code path of the self-signed hook (no cert, valid cert, outdated cert, outdated CA, mismatched expiry, shared CA with custom field names) using a real values store from [`testing/helpers`](../../testing/helpers).
- [`order_certificate_test.go`](./order_certificate_test.go) — JQ filter tests and a minimal config validation test.

For a functional end-to-end test, drive either hook through [`testing/framework`](../../testing/framework): seed the Secret in `KubeStateSet`, run the hook, and assert on the resulting values + cluster state.
