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

const (
	SnapshotKey = "apiservers"
)

var _ = registry.RegisterFunc(config, HandlerHook)

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
			Queue: "myqueue",
		},
	},
}

func HandlerHook(_ context.Context, input *pkg.HookInput) error {
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
