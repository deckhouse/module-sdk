package hookinfolder

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = registry.RegisterFunc(configKubernetesClient, HandlerKubernetesClient)

var configKubernetesClient = &pkg.HookConfig{}

func HandlerKubernetesClient(ctx context.Context, input *pkg.HookInput) error {
	k8sClient := input.DC.MustGetK8sClient()

	const (
		podNamespace = "test-namespace"
		podName      = "test-pod"
	)

	pod := new(corev1.Pod)
	err := k8sClient.Get(ctx, types.NamespacedName{Namespace: podNamespace, Name: podName}, pod)
	if err != nil {
		return fmt.Errorf("get pod: %w", err)
	}

	input.Logger.Info("pod", slog.String("name", pod.GetName()), slog.String("namespace", pod.GetNamespace()))

	return nil
}
