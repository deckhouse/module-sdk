package hookinfolder_test

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/mock"

	subfolder "dependency-example-module/subfolder"
)

var _ = Describe("k8s client hook example", func() {
	Context("refoncile func", func() {
		When("all services works correctly", func() {
			dc := mock.NewDependencyContainerMock(GinkgoT())
			dc.MustGetK8sClientMock.Set(func(_ ...pkg.KubernetesOption) pkg.KubernetesClient {
				return mock.NewKubernetesClientMock(GinkgoT()).GetMock.Set(func(_ context.Context, _ client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
					pod := obj.(*corev1.Pod)

					pod.Name = "found-pod"
					pod.Namespace = "found-ns"

					return nil
				})
			})

			var input = &pkg.HookInput{
				DC:     dc,
				Logger: log.NewNop(),
			}

			It("error has occurred", func() {
				err := subfolder.HandlerKubernetesClient(context.Background(), input)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		When("kubernetes client has an error", func() {
			dc := mock.NewDependencyContainerMock(GinkgoT())
			dc.MustGetK8sClientMock.Set(func(_ ...pkg.KubernetesOption) pkg.KubernetesClient {
				return mock.NewKubernetesClientMock(GinkgoT()).GetMock.Set(func(_ context.Context, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
					return errors.New("error")
				})
			})

			var input = &pkg.HookInput{
				DC:     dc,
				Logger: log.NewNop(),
			}

			It("error has occurred", func() {
				err := subfolder.HandlerKubernetesClient(context.Background(), input)
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(Equal(fmt.Errorf("get pod: %w", errors.New("error"))))
			})
		})
	})
})
