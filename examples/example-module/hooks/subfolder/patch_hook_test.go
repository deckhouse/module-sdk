package hookinfolder_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/pkg"
	testinghelpers "github.com/deckhouse/module-sdk/testing/helpers"

	subfolder "example-module/subfolder"
)

var _ = Describe("patch hook", func() {
	Context("HandlerHookPatch function", func() {
		var (
			patchCollector pkg.OutputPatchCollector

			buf   *bytes.Buffer
			input *pkg.HookInput
		)

		BeforeEach(func() {
			buf = bytes.NewBuffer([]byte{})

			input = &pkg.HookInput{
				PatchCollector: patchCollector,
				Logger: log.NewLogger(
					log.WithLevel(log.LevelDebug.Level()),
					log.WithOutput(buf),
					log.WithTimeFunc(func(_ time.Time) time.Time {
						parsedTime, err := time.Parse(time.DateTime, "2006-01-02 15:04:05")
						Expect(err).ShouldNot(HaveOccurred())
						return parsedTime
					}),
				),
			}
		})

		It("logs hello message and executes patch collector operations", func() {
			input.PatchCollector = testinghelpers.PreparePatchCollector(&testing.T{},
				testinghelpers.NewCreate(
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-first-pod",
							Namespace: "default",
						},
						Status: corev1.PodStatus{
							Phase: corev1.PodRunning,
						},
					}),
				testinghelpers.NewCreateOrUpdate(
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-second-pod",
							Namespace: "default",
						},
						Status: corev1.PodStatus{
							Phase: corev1.PodRunning,
						},
					},
				),
				testinghelpers.NewCreateIfNotExists(
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-third-pod",
							Namespace: "default",
						},
						Status: corev1.PodStatus{
							Phase: corev1.PodRunning,
						},
					},
				),
				testinghelpers.NewDelete(
					"v1", "Pod", "default", "my-first-pod",
				),
				testinghelpers.NewDeleteInBackground(
					"v1", "Pod", "default", "my-second-pod",
				),
				testinghelpers.NewDeleteNonCascading(
					"v1", "Pod", "default", "my-third-pod",
				),
				testinghelpers.NewPatchWithMerge(
					map[string]any{"/status": "newStatus"},
					"v1", "Pod", "default", "my-third-pod",
				),
			)

			// Execute the handler function
			err := subfolder.HandlerHookPatch(context.Background(), input)
			Expect(err).ShouldNot(HaveOccurred())

			// Verify log messages
			logs := strings.Split(buf.String(), "\n")
			Expect(logs[0]).To(ContainSubstring(`"level":"info","msg":"hello from patch hook"`))
		})
	})
})
