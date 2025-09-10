package hookinfolder_test

import (
	"context"
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/utils"
	"github.com/deckhouse/module-sdk/testing/mock"

	subfolder "example-module/subfolder"
)

var _ = Describe("values example", func() {
	Context("refoncile func", func() {
		When("all services works correctly", func() {
			values := mock.NewOutputPatchableValuesCollectorMock(GinkgoT())
			values.GetMock.When("some.path.to.field").Then(gjson.Result{Str: "str-value"})
			values.GetOkMock.When("some.path.to.field").Then(gjson.Result{Str: "str-value"}, true)
			values.GetPatchesMock.Return([]*utils.ValuesPatchOperation{
				{
					Op:    "add",
					Path:  "/some/path/to/field",
					Value: json.RawMessage(`{"name":"some-module"}`),
				},
			})
			values.GetRawMock.When("some.path.to.field.someInt").Then(float64(1))
			values.ExistsMock.When("some.path.to.field.str").Then(false)
			values.SetMock.When("some.path.to.field.str", "some_string")
			values.RemoveMock.When("some.path.to.field")
			values.ArrayCountMock.When("some.path.to.field.array").Then(10, nil)

			var input = &pkg.HookInput{
				Values: values,
				Logger: log.NewNop(),
			}

			It("reconcile func executed correctly", func() {
				err := subfolder.HandlerHookValues(context.Background(), input)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		When("get ok returns false", func() {
			values := mock.NewOutputPatchableValuesCollectorMock(GinkgoT())
			values.GetMock.When("some.path.to.field").Then(gjson.Result{Str: "str-value"})
			values.GetOkMock.When("some.path.to.field").Then(gjson.Result{Str: "str-value"}, false)
			values.ExistsMock.When("some.path.to.field.str").Then(false)
			values.SetMock.When("some.path.to.field.str", "some_string")
			values.RemoveMock.When("some.path.to.field")
			values.ArrayCountMock.When("some.path.to.field.array").Then(10, nil)

			var input = &pkg.HookInput{
				Values: values,
				Logger: log.NewNop(),
			}

			It("reconcile func executed correctly", func() {
				err := subfolder.HandlerHookValues(context.Background(), input)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		When("get raw geturns not number", func() {
			values := mock.NewOutputPatchableValuesCollectorMock(GinkgoT())
			values.GetMock.When("some.path.to.field").Then(gjson.Result{Str: "str-value"})
			values.GetOkMock.When("some.path.to.field").Then(gjson.Result{Str: "str-value"}, true)
			values.GetPatchesMock.Return([]*utils.ValuesPatchOperation{
				{
					Op:    "add",
					Path:  "/some/path/to/field",
					Value: json.RawMessage(`{"name":"some-module"}`),
				},
			})
			values.GetRawMock.When("some.path.to.field.someInt").Then("not number")
			values.ExistsMock.When("some.path.to.field.str").Then(false)
			values.SetMock.When("some.path.to.field.str", "some_string")
			values.RemoveMock.When("some.path.to.field")
			values.ArrayCountMock.When("some.path.to.field.array").Then(10, nil)

			var input = &pkg.HookInput{
				Values: values,
				Logger: log.NewNop(),
			}

			It("reconcile func executed correctly", func() {
				err := subfolder.HandlerHookValues(context.Background(), input)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		When("array count returns error", func() {
			values := mock.NewOutputPatchableValuesCollectorMock(GinkgoT())
			values.GetMock.When("some.path.to.field").Then(gjson.Result{Str: "str-value"})
			values.GetOkMock.When("some.path.to.field").Then(gjson.Result{Str: "str-value"}, true)
			values.GetPatchesMock.Return([]*utils.ValuesPatchOperation{
				{
					Op:    "add",
					Path:  "/some/path/to/field",
					Value: json.RawMessage(`{"name":"some-module"}`),
				},
			})
			values.GetRawMock.When("some.path.to.field.someInt").Then(float64(1))
			values.ExistsMock.When("some.path.to.field.str").Then(false)
			values.SetMock.When("some.path.to.field.str", "some_string")
			values.RemoveMock.When("some.path.to.field")
			values.ArrayCountMock.When("some.path.to.field.array").Then(0, errors.New("error"))

			var input = &pkg.HookInput{
				Values: values,
				Logger: log.NewNop(),
			}

			It("reconcile func executed correctly", func() {
				err := subfolder.HandlerHookValues(context.Background(), input)
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(Equal(errors.New("error")))
			})
		})
	})
})
