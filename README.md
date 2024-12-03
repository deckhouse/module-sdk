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

### Available transports

#### Files in

Example config values file
```json
{
    "global":{
        "modules":{
            "publicDomainTemplate":"%s.com"
            }
    },
    "some-module":{}
}
```

Example binding contexts file
```json
[
	{
	"binding": "node_roles",
	"groupName": "policy",
	"snapshots": {
		"node_roles": [
		{
			"object": {
				"apiVersion": "v1",
				"metadata": {
					"name": "test-object"
				}
			}
		},
		{
			"filterResult": {
				"apiVersion": "v1",
				"metadata": {
					"name": "test-filter-result"
				}
			}
		}
		]
	},
	"type": "Group"
	}
]
```

#### Files out

Correspond with struct tags of entities with WriteOutput(w io.Writer) error method

### Work sequence

#### Deckhouse register process
1) To register your hooks, add them to import section in main package like in examples
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