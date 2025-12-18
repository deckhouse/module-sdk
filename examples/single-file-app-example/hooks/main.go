package main

import (
	"context"
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/app"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

const (
	SnapshotKey = "apiservers"
)

var _ = registry.RegisterFunc(config, Handle)

var config = &pkg.HookConfig{
	Kubernetes: []pkg.KubernetesConfig{
		{
			Name:       SnapshotKey,
			APIVersion: "metav1",
			Kind:       "Pod",
			NamespaceSelector: &pkg.NamespaceSelector{
				NameSelector: &pkg.NameSelector{
					MatchNames: []string{"kube-system"},
				},
			},
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"component": "kube-apiserver"},
			},
			JqFilter: ".metadata.name",
		},
	},
}

func Handle(_ context.Context, input *pkg.ApplicationHookInput) error {
	podNames, err := objectpatch.UnmarshalToStruct[string](input.Snapshots, SnapshotKey)
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
