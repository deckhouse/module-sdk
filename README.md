# Module SDK
SDK to easy compile your hooks as a binary and integrate with addon operator

## Usage

### One file example
This file must be in 'hooks/' folder to build binary (see examples for correct layout)

```go
package main

import (
  "context"
  "log/slog"

  "github.com/deckhouse/module-sdk/pkg"
  "github.com/deckhouse/module-sdk/pkg/app"
  objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
  "github.com/deckhouse/module-sdk/pkg/registry"
  v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = registry.RegisterFunc(config, handlerHook)

var config = &pkg.HookConfig{
  Kubernetes: []pkg.KubernetesConfig{
    {
      Name:       "apiservers",
      APIVersion: "v1",
      Kind:       "Pod",
      NamespaceSelector: &pkg.NamespaceSelector{
        NameSelector: &pkg.NameSelector{
          MatchNames: []string{"kube-system"},
        },
      },
      LabelSelector: &v1.LabelSelector{
        MatchLabels: map[string]string{"component": "kube-apiserver"},
      },
      JqFilter: ".metadata.name",
    },
  },
}

func handlerHook(_ context.Context, input *pkg.HookInput) error {
  podNames, err := objectpatch.UnmarshalToStruct[string](input.Snapshots, "apiservers")
  if err != nil {
    return err
  }

  input.Logger.Info("found apiserver pods", slog.Any("podNames", podNames))

  input.Values.Set("test.internal.apiServers", podNames)

  return nil
}

func main() {
  app.Run()
}
```

More examples you can find [here](./examples)

## Adding Readiness Probes

Readiness probes let your module report its ready status to the Deckhouse addon-operator.

### How to Add a Readiness Probe

1. **Create a Readiness Function**

   Create a function that checks if your module is ready. This function should return `nil` if everything is okay or an error if the module is not ready.

   ```go
   func checkReadiness(ctx context.Context) error {
     // Perform your ready checks here
     // Return nil if ready, or error if not ready
     return nil
   }
   ```

2. **Register Your Readiness Probe**

   Register your readiness function when initializing your application:

   ```go
   func main() {
     readinessConfig := app.ReadinessConfig{
        // you can override it with environment variable
        // default: 15
        IntervalInSeconds: 12,
        ProbeFunc: checkReadiness,
     }
     
     app.RunWithReadinessProbe(readinessConfig)
   }
   ```

3. **Example: HTTP Endpoint Check**

   Here's a complete example that checks if an HTTP endpoint is available:

   ```go
   package main

   import (
     "context"
     "fmt"
     "net/http"
     "time"

     "github.com/deckhouse/module-sdk/pkg/app"
     "github.com/deckhouse/module-sdk/pkg/registry"
   )

   var _ = registry.RegisterFunc(config, handlerHook)

   // Your regular hook config and handler here

   func checkAPIEndpoint(ctx context.Context) error {
	   client := input.DC.GetHTTPClient()
     
     req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://api.example.com/readyz", nil)
     if err != nil {
       return fmt.Errorf("failed to create request: %w", err)
     }
     
     resp, err := client.Do(req)
     if err != nil {
       return fmt.Errorf("API endpoint unreachable: %w", err)
     }
     defer resp.Body.Close()
     
     if resp.StatusCode != http.StatusOK {
       return fmt.Errorf("API endpoint returned non-OK status: %d", resp.StatusCode)
     }
     
     return nil
   }

   func main() {
     readinessConfig := app.ReadinessConfig{
       ProbeFunc: checkAPIEndpoint,
     }
     
     app.RunWithReadinessProbe(readinessConfig)
   }
   ```

### Behavior

- When your readiness function succeeds (returns `nil`), the module status changes to `Ready`
- When it fails (returns an error), the status changes to `Reconciling` with the error message
- The module resource's `IsReady` condition is updated to reflect the current state
- This lets other components in Deckhouse know when your module is operational

### Configuration Options

You can configure the readiness probe using environment variables:

| Variable | Description | Default |
| --- | --- | --- |
| `READINESS_INTERVAL_IN_SECONDS` | How often to check readiness (in seconds) | None |
| `MODULE_NAME` | Module name used in readiness reporting | default-module |

## Testing

If you want to test your JQ filter, you can use JQ helper like in example [here](./pkg/jq/jq_test.go)

## For deckhouse developers

### Environment variables

| Parameter | Required | Default value | Description |
| --- | --- | --- | --- |
| BINDING_CONTEXT_PATH |  | in/binding_context.json | Path to binding context file |
| VALUES_PATH |  | in/values_path.json | Path to values file |
| CONFIG_VALUES_PATH |  | in/config_values_path.json | Path to config values file |
| METRICS_PATH |  | out/metrics.json | Path to metrics file |
| KUBERNETES_PATCH_PATH |  | out/kubernetes.json | Path to kubernetes patch file |
| VALUES_JSON_PATCH_PATH |  | out/values.json | Path to values patch file |
| CONFIG_VALUES_JSON_PATCH_PATH |  | out/config_values.json | Path to config values patch file |
| HOOK_CONFIG_PATH |  | out/hook_config.json | Path to dump hook configurations in file |
| CREATE_FILES |  | false | Allow hook to create files by himself (by default, waiting for addon operator to create) |
| MODULE_NAME |  | default-module | Name of the module, hooks align |
| READINESS_INTERVAL_IN_SECONDS |  |  | Interval in seconds for module readiness checks (override user values) |
| LOG_LEVEL |  | FATAL | Log level (suppressed by default) |

### Work sequence

#### Deckhouse register process
1) To register your hooks, add them to import section in main package like in [examples](./examples)
2) Compile your binary and deliver to "hooks" folder in Deckhouse
3) Addon operator finds it automatically and register all your hooks in binary, corresponding with your HookConfigs
4) When addon operator has a reason, it calls hook in your binary 
5) After executing hook, addon operator process hook output

#### Calling hook
1) Addon operator create temporary files for input and output data (see ENV for examples)
2) Addon operator execute hook with corresponding ID and ENV variables pointed to files
3) Hook reads all files and pass incoming data in HookInput
4) Hook executes and write all resulting data from collectors contained in HookInput
5) Addon operator reads info from temporary output files
