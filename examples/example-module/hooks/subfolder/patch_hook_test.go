package hookinfolder_test

import (
	"context"
	subfolder "example-module/subfolder"

	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/deckhouse/module-sdk/pkg"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	"github.com/deckhouse/module-sdk/testing/mock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Patch hook", func() {
	Context("Patch collector operations", func() {
		When("all operations work correctly", func() {
			patchCollector := mock.NewPatchCollectorMock(GinkgoT())

			// Setup pod expectations
			firstPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-first-pod",
					Namespace: "default",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			}

			secondPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-second-pod",
					Namespace: "default",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			}

			thirdPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-third-pod",
					Namespace: "default",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			}

			// Mock Create operations
			patchCollector.CreateMock.Expect(firstPod, mock.Any())
			patchCollector.CreateOrUpdateMock.Expect(secondPod, mock.Any())
			patchCollector.CreateIfNotExistsMock.Expect(thirdPod, mock.Any())

			// Mock Delete operations
			patchCollector.DeleteMock.Expect(firstPod.APIVersion, "Pod", "default", "my-first-pod", mock.Any())
			patchCollector.DeleteInBackgroundMock.Expect(secondPod.APIVersion, "Pod", "default", "my-second-pod", mock.Any())
			patchCollector.DeleteNonCascadingMock.Expect(thirdPod.APIVersion, "Pod", "default", "my-third-pod", mock.Any())

			// Mock MergePatch operation
			expectedPatch := map[string]interface{}{
				"/status": "newStatus",
			}
			patchCollector.MergePatchMock.Expect(expectedPatch, thirdPod.APIVersion, "Pod", "default", "my-third-pod", mock.Any())

			var input = &pkg.HookInput{
				PatchCollector: patchCollector,
				Logger:         log.NewNop(),
			}

			It("executes all patch operations successfully", func() {
				err := subfolder.HandlerHookPatch(context.Background(), input)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		When("testing creation operations only", func() {
			patchCollector := mock.NewPatchCollectorMock(GinkgoT())

			// Mock Create operations with specific matchers
			patchCollector.CreateMock.Set(func(obj interface{}, options ...objectpatch.CreateOption) {
				pod, ok := obj.(*corev1.Pod)
				Expect(ok).To(BeTrue())
				Expect(pod.Name).To(Equal("my-first-pod"))
			})

			patchCollector.CreateOrUpdateMock.Set(func(obj interface{}, options ...objectpatch.CreateOption) {
				pod, ok := obj.(*corev1.Pod)
				Expect(ok).To(BeTrue())
				Expect(pod.Name).To(Equal("my-second-pod"))
			})

			patchCollector.CreateIfNotExistsMock.Set(func(obj interface{}, options ...objectpatch.CreateOption) {
				pod, ok := obj.(*corev1.Pod)
				Expect(ok).To(BeTrue())
				Expect(pod.Name).To(Equal("my-third-pod"))
			})

			// Mock other operations that will be called
			patchCollector.DeleteMock.Return()
			patchCollector.DeleteInBackgroundMock.Return()
			patchCollector.DeleteNonCascadingMock.Return()
			patchCollector.MergePatchMock.Return()

			var input = &pkg.HookInput{
				PatchCollector: patchCollector,
				Logger:         log.NewNop(),
			}

			It("correctly creates all pods", func() {
				err := subfolder.HandlerHookPatch(context.Background(), input)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		When("testing deletion operations only", func() {
			patchCollector := mock.NewPatchCollectorMock(GinkgoT())

			// Mock other operations that will be called
			patchCollector.CreateMock.Return()
			patchCollector.CreateOrUpdateMock.Return()
			patchCollector.CreateIfNotExistsMock.Return()
			patchCollector.MergePatchMock.Return()

			// Mock Delete operations with specific expectations
			patchCollector.DeleteMock.Expect("", "Pod", "default", "my-first-pod", mock.Any())
			patchCollector.DeleteInBackgroundMock.Expect("", "Pod", "default", "my-second-pod", mock.Any())
			patchCollector.DeleteNonCascadingMock.Expect("", "Pod", "default", "my-third-pod", mock.Any())

			var input = &pkg.HookInput{
				PatchCollector: patchCollector,
				Logger:         log.NewNop(),
			}

			It("correctly deletes all pods", func() {
				err := subfolder.HandlerHookPatch(context.Background(), input)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		When("testing merge patch operation only", func() {
			patchCollector := mock.NewPatchCollectorMock(GinkgoT())

			// Mock other operations that will be called
			patchCollector.CreateMock.Return()
			patchCollector.CreateOrUpdateMock.Return()
			patchCollector.CreateIfNotExistsMock.Return()
			patchCollector.DeleteMock.Return()
			patchCollector.DeleteInBackgroundMock.Return()
			patchCollector.DeleteNonCascadingMock.Return()

			// Expected patch
			expectedPatch := map[string]interface{}{
				"/status": "newStatus",
			}

			// Mock MergePatch operation with specific expectations
			patchCollector.MergePatchMock.Expect(expectedPatch, "", "Pod", "default", "my-third-pod", mock.Any())

			var input = &pkg.HookInput{
				PatchCollector: patchCollector,
				Logger:         log.NewNop(),
			}

			It("correctly applies merge patch", func() {
				err := subfolder.HandlerHookPatch(context.Background(), input)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})
})
