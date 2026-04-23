package main_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/mock"

	singlefileappexample "singlefileappexample"
)

const (
	firstSnapshot  = "one"
	secondSnapshot = "two"
)

var _ = Describe("handle hook single file example", func() {
	Context("settings gate closed", func() {
		settings := mock.NewOutputPatchableValuesCollectorMock(GinkgoT())
		settings.GetOkMock.When("apiServersDiscovery.enabled").Then(gjson.Result{}, false)

		values := mock.NewOutputPatchableValuesCollectorMock(GinkgoT())

		var input = &pkg.ApplicationHookInput{
			Values:   values,
			Settings: settings,
			Logger:   log.NewNop(),
		}

		It("does not touch values when the gate is closed", func() {
			err := singlefileappexample.Handle(context.Background(), input)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("settings gate open", func() {
		snapshots := mock.NewSnapshotsMock(GinkgoT())
		snapshots.GetMock.When(singlefileappexample.SnapshotKey).Then(
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

		settings := mock.NewOutputPatchableValuesCollectorMock(GinkgoT())
		settings.GetOkMock.When("apiServersDiscovery.enabled").Then(gjson.Result{Type: gjson.True}, true)

		values := mock.NewOutputPatchableValuesCollectorMock(GinkgoT())
		values.SetMock.When("test.internal.apiServers", []string{firstSnapshot, secondSnapshot})

		var input = &pkg.ApplicationHookInput{
			Snapshots: snapshots,
			Values:    values,
			Settings:  settings,
			Logger:    log.NewNop(),
		}

		It("writes discovered pods into values when the gate is open", func() {
			err := singlefileappexample.Handle(context.Background(), input)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
