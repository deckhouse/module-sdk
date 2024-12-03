# Examples

[Module hook basic example](https://github.com/deckhouse/module-sdk/tree/main/examples/basic-example-module)

[Module hook single file example](https://github.com/deckhouse/module-sdk/tree/main/examples/single-file-example)

[Module hook example](https://github.com/deckhouse/module-sdk/tree/main/examples/example-module)

[Module hook example with dependency container](https://github.com/deckhouse/module-sdk/tree/main/examples/dependency-example-module)

[Dockerfile and Makefile for building](https://github.com/deckhouse/module-sdk/tree/main/examples/scripts)

### One file example
This file must be in 'hooks/' folder to build binary

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