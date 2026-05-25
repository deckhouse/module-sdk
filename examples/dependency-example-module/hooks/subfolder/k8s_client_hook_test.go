package hookinfolder_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/helpers"
	"github.com/deckhouse/module-sdk/testing/mock"

	subfolder "dependency-example-module/subfolder"
)

// k8sDC builds a DependencyContainerMock whose MustGetK8sClient returns
// a KubernetesClient with the provided Get behaviour.
func k8sDC(t *testing.T, getFn func(context.Context, client.ObjectKey, client.Object, ...client.GetOption) error) pkg.DependencyContainer {
	t.Helper()
	dc := mock.NewDependencyContainerMock(t)
	dc.MustGetK8sClientMock.Set(func(_ ...pkg.KubernetesOption) pkg.KubernetesClient {
		return mock.NewKubernetesClientMock(t).GetMock.Set(getFn)
	})
	return dc
}

func TestHandlerKubernetesClient_HappyPath(t *testing.T) {
	dc := k8sDC(t, func(_ context.Context, _ client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
		pod := obj.(*corev1.Pod)
		pod.Name = "found-pod"
		pod.Namespace = "found-ns"
		return nil
	})

	in := helpers.NewInputBuilder(t).WithDependencyContainer(dc).Build()
	require.NoError(t, subfolder.HandlerKubernetesClient(context.Background(), in))
}

func TestHandlerKubernetesClient_ReturnsWrappedError(t *testing.T) {
	wantErr := errors.New("boom")
	dc := k8sDC(t, func(_ context.Context, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
		return wantErr
	})

	in := helpers.NewInputBuilder(t).WithDependencyContainer(dc).Build()

	err := subfolder.HandlerKubernetesClient(context.Background(), in)
	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
	assert.ErrorContains(t, err, "get pod:")
}
