# Examples

Each example is a standalone Go module with its own `go.mod` and tests, so it doubles as documentation for one specific feature of the Module SDK.

## Module hooks

| Example | What it shows |
| --- | --- |
| [`basic-example-module`](./basic-example-module) | The simplest possible layout: a `hooks/` folder, a single hook, a single binary. |
| [`single-file-example`](./single-file-example) | Single-file module hook plus tests using both [`testing/helpers`](../testing/helpers) and [`testing/framework`](../testing/framework). |
| [`example-module`](./example-module) | A richer module with several hooks (snapshots, patches, values, metrics) and tests for each. |
| [`dependency-example-module`](./dependency-example-module) | Hooks that use the `DependencyContainer` (HTTP client, K8s client, registry client), with mock-based and framework-level tests. |
| [`common-hooks`](./common-hooks) | Real-world usage of the reusable hooks under [`common-hooks/`](../common-hooks) (e.g. `tls-certificate`). |

## Application hooks

| Example | What it shows |
| --- | --- |
| [`single-file-app-example`](./single-file-app-example) | A minimal `pkg.ApplicationHookInput` hook, including a settings gate and a snapshot binding. |
| [`settings-check`](./settings-check) | Validating module configuration values via `pkg/settingscheck`. |

## Build & deploy

| Example | What it shows |
| --- | --- |
| [`scripts`](./scripts) | Reference `Dockerfile` and `Makefile` for building and shipping a hook binary. |

---

## Running the examples

Every example is a self-contained module with a `replace` directive pointing back at this repo, so you can do:

```sh
cd examples/example-module/hooks
go test ./...
```

The repository-wide `make examples` target iterates over every example module.

## Test layering

The example suites intentionally mix testing styles, so you can copy whichever fits your situation:

- **Unit tests** with `testing/helpers.InputBuilder` + real values stores;
- **Mock-based tests** for error-injection edge cases;
- **Functional tests** (`*_framework_test.go`) driving the hook against a fake Kubernetes cluster via `testing/framework`.

For the strategy behind that mix, see [`TESTING.md`](../TESTING.md).
