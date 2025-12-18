package main_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/mock"

	singlefileexample "singlefileexample"
)

const (
	firstSnapshot  = "one"
	secondSnapshot = "two"
)

var _ = Describe("handle hook single file example", func() {
	snapshots := mock.NewSnapshotsMock(GinkgoT())
	snapshots.GetMock.When(singlefileexample.SnapshotKey).Then(
		[]pkg.Snapshot{
			mock.NewSnapshotMock(GinkgoT()).UnmarshalToMock.Set(func(v any) error {
				str := v.(*string)
				*str = firstSnapshot

				return nil
			}),
			mock.NewSnapshotMock(GinkgoT()).UnmarshalToMock.Set(func(v any) error {
				str := v.(*string)
				*str = secondSnapshot

				return nil
			}),
		},
	)

	values := mock.NewOutputPatchableValuesCollectorMock(GinkgoT())
	values.SetMock.When("test.internal.apiServers", []string{firstSnapshot, secondSnapshot})

	var input = &pkg.ApplicationHookInput{
		Snapshots: snapshots,
		Values:    values,
		Logger:    log.NewNop(),
	}

	Context("reconcile func", func() {
		It("reconcile func executed correctly", func() {
			err := singlefileexample.Handle(context.Background(), input)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
