package hookinfolder_test

import (
	"context"
	"errors"
	"fmt"

	subfolder "dependency-example-module/subfolder"

	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/mock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("k8s client hook example", func() {
	Context("refoncile func", func() {
		When("all services works correctly", func() {
			dc := mock.NewDependencyContainerMock(GinkgoT())
			dc.MustGetK8sClientMock.Set(func(options ...pkg.KubernetesOption) (k1 pkg.KubernetesClient) {
				k8sClient := mock.NewKubernetesClientMock(GinkgoT()).SchemeMock.Set(func() (sp1 *runtime.Scheme) {
					return runtime.NewScheme()
				})

				k8sClient.GetMock.Set(func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) (err error) {
					pod := obj.(*corev1.Pod)

					pod.Name = "found-pod"
					pod.Namespace = "found-ns"

					return nil
				})

				return k8sClient
			})

			var input = &pkg.HookInput{
				DC:     dc,
				Logger: log.NewNop(),
			}

			It("error has occured", func() {
				err := subfolder.HandlerKubernetesClient(context.Background(), input)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		When("kubernetes client has an error", func() {
			dc := mock.NewDependencyContainerMock(GinkgoT())
			dc.MustGetK8sClientMock.Set(func(options ...pkg.KubernetesOption) (k1 pkg.KubernetesClient) {
				k8sClient := mock.NewKubernetesClientMock(GinkgoT()).SchemeMock.Set(func() (sp1 *runtime.Scheme) {
					return runtime.NewScheme()
				})

				k8sClient.GetMock.Set(func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) (err error) {
					return errors.New("error")
				})

				return k8sClient
			})

			var input = &pkg.HookInput{
				DC:     dc,
				Logger: log.NewNop(),
			}

			It("error has occured", func() {
				err := subfolder.HandlerKubernetesClient(context.Background(), input)
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(Equal(fmt.Errorf("get pod: %w", errors.New("error"))))
			})
		})
	})
})
