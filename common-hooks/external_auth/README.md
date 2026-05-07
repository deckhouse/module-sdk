# `external_auth`

Module hook that wires external authentication settings into the module's values, taking into account whether the cluster has the `user-authn` module enabled and whether the user supplied an explicit configuration.

## Decision matrix

The hook runs `OnBeforeHelm` (order `9`) and computes the final value at the configured `Settings.ExternalAuthPath`:

| `user-authn` enabled? | User provided `ExternalAuthPath` in config? | Resulting values |
| --- | --- | --- |
| no | no | `ExternalAuthPath` removed from values; `DexAuthenticatorEnabledPath` removed. |
| no | yes | `ExternalAuthPath` set to user-provided value; `DexAuthenticatorEnabledPath` removed. |
| yes | no | `ExternalAuthPath` set to the Dex template (`Settings.DexExternalAuth`, with `%CLUSTER_DOMAIN%` substituted); `DexAuthenticatorEnabledPath` set to `true`. |
| yes | yes | `ExternalAuthPath` set to user-provided value; `DexAuthenticatorEnabledPath` removed (the user opted out of the in-module authenticator). |

In other words: **the user always wins**, and the in-module Dex authenticator is only enabled when (a) `user-authn` is enabled and (b) the user did not supply their own auth.

## Settings

```go
type Settings struct {
    // Where in values the resulting auth block lives.
    ExternalAuthPath string

    // Where to set the boolean flag that gates the in-module DexAuthenticator.
    DexAuthenticatorEnabledPath string

    // What to write to ExternalAuthPath when Dex is the source of truth.
    DexExternalAuth ExternalAuth
}

type ExternalAuth struct {
    AuthURL         string  // may contain "%CLUSTER_DOMAIN%"
    AuthSignInURL   string
    UseBearerTokens *bool
}
```

`ExternalAuth.AuthURLWithClusterDomain(input)` substitutes `%CLUSTER_DOMAIN%` with `global.discovery.clusterDomain`.

## Usage

```go
package hooks

import (
    extauth "github.com/deckhouse/module-sdk/common-hooks/external_auth"
)

var useBearer = true

var _ = extauth.RegisterHook(extauth.Settings{
    ExternalAuthPath:            "myModule.auth.externalAuthentication",
    DexAuthenticatorEnabledPath: "myModule.internal.dexAuthenticatorEnabled",
    DexExternalAuth: extauth.ExternalAuth{
        AuthURL:         "https://dex.%CLUSTER_DOMAIN%/auth",
        AuthSignInURL:   "https://signin.%CLUSTER_DOMAIN%/",
        UseBearerTokens: &useBearer,
    },
})
```

## Hook configuration

- **Trigger:** `OnBeforeHelm` with `Order: 9`.
- **No Kubernetes bindings:** the hook only inspects values.
- **Reads:**
    - `global.enabledModules` (uses `pkg/utils/set` to check for `user-authn`),
    - `global.discovery.clusterDomain` (for `%CLUSTER_DOMAIN%` expansion),
    - `<Settings.ExternalAuthPath>` from `ConfigValues` (the user override, if any).

## Testing

The hook is a pure values-mutating function, which makes it ideal for unit tests with [`testing/helpers`](../../testing/helpers):

```go
in := helpers.NewInputBuilder(t).
    WithValuesJSON(`{"global":{"enabledModules":["user-authn"]}}`).
    WithConfigValuesJSON(`{}`).
    Build()
```

Then call the hook (export the handler from your test code or use `RegisterHook` with a custom registry) and assert on `in.Values.GetPatches()`.
