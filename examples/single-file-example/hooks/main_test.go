package main_test

import (
	"context"

	singlefileexample "singlefileexample"

	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/mock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	firstSnapshot  = "one"
	secondSnapshot = "two"
)

var _ = Describe("handle hook single file example", func() {
	snapshots := mock.NewSnapshotsMock(GinkgoT())
	snapshots.GetMock.When(singlefileexample.SnapshotKey).Then(
		[]pkg.Snapshot{
			mock.NewSnapshotMock(GinkgoT()).UnmarhalToMock.Set(func(v any) (err error) {
				str := v.(*string)
				*str = firstSnapshot

				return nil
			}),
			mock.NewSnapshotMock(GinkgoT()).UnmarhalToMock.Set(func(v any) (err error) {
				str := v.(*string)
				*str = secondSnapshot

				return nil
			}),
		},
	)

	values := mock.NewPatchableValuesCollectorMock(GinkgoT())
	values.SetMock.When("test.internal.apiServers", []string{firstSnapshot, secondSnapshot})

	var input = &pkg.HookInput{
		Snapshots: snapshots,
		Values:    values,
		Logger:    log.NewNop(),
	}

	Context("refoncile func", func() {
		It("reconcile func executed correctly", func() {
			err := singlefileexample.HandlerHook(context.Background(), input)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
