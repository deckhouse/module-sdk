package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

func Handle(_ context.Context, input *pkg.HookInput) error {
	podNames, err := objectpatch.UnmarshalToStruct[string](input.Snapshots, "apiservers")
	if err != nil {
		return err
	}

	input.Logger.Info("found apiserver pods", slog.Any("podNames", podNames))

	input.Values.Set("test.internal.apiServers", podNames)

	return nil
}

func ReadinessFunc(ctx context.Context, input *pkg.HookInput) error {
	input.Logger.Info("start user logic for readiness probe")

	c := input.DC.GetHTTPClient()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1/readyz", nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	input.Logger.Debug("readiness probe done successfully", slog.Any("body", string(respBody)))

	return nil
}

func main() {
	readinessConfig := &app.ReadinessConfig{
		IntervalInSeconds: 12,
		ProbeFunc:         ReadinessFunc,
	}

	app.Run(app.WithReadiness(readinessConfig))
}
