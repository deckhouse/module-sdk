# Module SDK
SDK to easy compile your hooks as a binary and integrate with addon operator

### Using

See [examples](https://github.com/deckhouse/module-sdk/tree/main/examples)

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
| LOG_LEVEL |  | FATAL | Log level (suppressed by default) |