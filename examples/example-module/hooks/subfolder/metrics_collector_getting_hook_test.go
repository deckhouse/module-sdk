package hookinfolder_test

import (
	"context"

	subfolder "example-module/subfolder"

	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/mock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("metrics collector example", func() {
	collector := mock.NewMetricsCollectorMock(GinkgoT())

	collector.AddMock.Set(func(name string, value float64, labels map[string]string, opts ...pkg.Option) {
		Expect(name).Should(Equal("stub-add-metric"))
		Expect(value).Should(Equal(float64(1)))
		Expect(labels).Should(Equal(map[string]string{"node_found": "node_name"}))
	})

	collector.SetMock.Set(func(name string, value float64, labels map[string]string, opts ...pkg.Option) {
		Expect(name).Should(Equal("stub-set-metric"))
		Expect(value).Should(Equal(float64(1)))
		Expect(labels).Should(Equal(map[string]string{"node_found": "node_name"}))
	})

	collector.IncMock.Set(func(name string, labels map[string]string, opts ...pkg.Option) {
		Expect(name).Should(Equal("stub-inc-metric"))
		Expect(labels).Should(Equal(map[string]string{"node_found": "node_name"}))
	})

	var input = &pkg.HookInput{
		MetricsCollector: collector,
		Logger:           log.NewNop(),
	}

	Context("refoncile func", func() {
		It("reconcile func executed correctly", func() {
			err := subfolder.HandlerHookMetricsCollector(context.Background(), input)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
