package main

import (
	"context"
	"log/slog"

	"github.com/deckhouse/module-sdk/pkg/settingscheck"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/app"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

const (
	SnapshotKey = "apiservers"
)

var _ = registry.RegisterAppFunc(config, Handle)

var config = &pkg.HookConfig{
	Kubernetes: []pkg.KubernetesConfig{
		{
			Name:       SnapshotKey,
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

func Handle(_ context.Context, input *pkg.ApplicationHookInput) error {
	podNames, err := objectpatch.UnmarshalToStruct[string](input.Snapshots, SnapshotKey)
	if err != nil {
		return err
	}

	input.Logger.Info("found apiserver pods", slog.Any("podNames", podNames))

	input.Values.Set("test.internal.apiServers", podNames)

	return nil
}

func check(_ context.Context, input settingscheck.Input) settingscheck.Result {
	replicas := input.Settings.Get("replicas").Int()
	if replicas == 0 {
		return settingscheck.Reject("replicas cannot be 0")
	}

	var warnings []string
	if replicas == 2 {
		warnings = append(warnings, "replicas cannot be greater than 3")
	}

	if replicas > 3 {
		return settingscheck.Reject("replicas cannot be greater than 3")
	}

	return settingscheck.Allow(warnings...)
}

func main() {
	app.Run(app.WithSettingsCheck(check))
}
