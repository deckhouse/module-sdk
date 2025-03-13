package hookinfolder_test

import (
	"bytes"
	"context"
	"strings"
	"time"

	subfolder "example-module/subfolder"

	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/mock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("patch hook", func() {
	Context("HandlerHookPatch function", func() {
		var (
			patchCollector *mock.PatchCollectorMock
			buf            *bytes.Buffer
			input          *pkg.HookInput
		)

		BeforeEach(func() {
			patchCollector = mock.NewPatchCollectorMock(GinkgoT())
			buf = bytes.NewBuffer([]byte{})

			input = &pkg.HookInput{
				PatchCollector: patchCollector,
				Logger: log.NewLogger(log.Options{
					Level:  log.LevelDebug.Level(),
					Output: buf,
					TimeFunc: func(_ time.Time) time.Time {
						parsedTime, err := time.Parse(time.DateTime, "2006-01-02 15:04:05")
						Expect(err).ShouldNot(HaveOccurred())
						return parsedTime
					},
				}),
			}
		})

		It("logs hello message and executes patch collector operations", func() {
			// Set expectations for Create
			patchCollector.CreateMock.Set(func(obj any) {
				pod, ok := obj.(*corev1.Pod)
				Expect(ok).To(BeTrue())
				Expect(pod.Name).To(Equal("my-first-pod"))
				Expect(pod.Namespace).To(Equal("default"))
				Expect(pod.Status.Phase).To(Equal(corev1.PodRunning))
			})

			// Set expectations for CreateOrUpdate
			patchCollector.CreateOrUpdateMock.Set(func(obj any) {
				pod, ok := obj.(*corev1.Pod)
				Expect(ok).To(BeTrue())
				Expect(pod.Name).To(Equal("my-second-pod"))
				Expect(pod.Namespace).To(Equal("default"))
				Expect(pod.Status.Phase).To(Equal(corev1.PodRunning))
			})

			// Set expectations for CreateIfNotExists
			patchCollector.CreateIfNotExistsMock.Set(func(obj any) {
				pod, ok := obj.(*corev1.Pod)
				Expect(ok).To(BeTrue())
				Expect(pod.Name).To(Equal("my-third-pod"))
				Expect(pod.Namespace).To(Equal("default"))
				Expect(pod.Status.Phase).To(Equal(corev1.PodRunning))
			})

			// Set expectations for Delete
			patchCollector.DeleteMock.Set(func(apiVersion, kind, namespace, name string) {
				Expect(apiVersion).To(Equal("v1"))
				Expect(kind).To(Equal("Pod"))
				Expect(namespace).To(Equal("default"))
				Expect(name).To(Equal("my-first-pod"))
			})

			// Set expectations for DeleteInBackground
			patchCollector.DeleteInBackgroundMock.Set(func(apiVersion, kind, namespace, name string) {
				Expect(apiVersion).To(Equal("v1"))
				Expect(kind).To(Equal("Pod"))
				Expect(namespace).To(Equal("default"))
				Expect(name).To(Equal("my-second-pod"))
			})

			// Set expectations for DeleteNonCascading
			patchCollector.DeleteNonCascadingMock.Set(func(apiVersion, kind, namespace, name string) {
				Expect(apiVersion).To(Equal("v1"))
				Expect(kind).To(Equal("Pod"))
				Expect(namespace).To(Equal("default"))
				Expect(name).To(Equal("my-third-pod"))
			})

			// Set expectations for MergePatch
			patchCollector.MergePatchMock.Set(func(patch any, apiVersion, kind, namespace, name string, opts ...pkg.PatchCollectorOption) {
				patchMap, ok := patch.(map[string]any)
				Expect(ok).To(BeTrue())
				Expect(patchMap).To(HaveKeyWithValue("/status", "newStatus"))
				Expect(apiVersion).To(Equal("v1"))
				Expect(kind).To(Equal("Pod"))
				Expect(namespace).To(Equal("default"))
				Expect(name).To(Equal("my-third-pod"))
				Expect(len(opts)).To(Equal(2))
			})

			// Execute the handler function
			err := subfolder.HandlerHookPatch(context.Background(), input)
			Expect(err).ShouldNot(HaveOccurred())

			// Verify log messages
			logs := strings.Split(buf.String(), "\n")
			Expect(logs[0]).To(ContainSubstring(`"level":"info","msg":"hello from patch hook"`))
		})
	})
})
