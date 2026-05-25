package hookinfolder_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/helpers"
	"github.com/deckhouse/module-sdk/testing/mock"

	subfolder "example-module/subfolder"
)

// The values hook reads a few paths and writes one. Where possible we
// drive it with a real PatchableValues store (via helpers.NewValuesFromJSON);
// only the deliberately error-injecting cases keep the mock.

func TestHandlerHookValues_HappyPath(t *testing.T) {
	values := helpers.NewValuesFromJSON(`{
        "some": {
            "path": {
                "to": {
                    "field": {
                        "someInt": 1,
                        "array":   [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
                    }
                }
            }
        }
    }`)

	in := helpers.NewInputBuilder(t).WithValues(values).Build()

	require.NoError(t, subfolder.HandlerHookValues(context.Background(), in))

	patches := in.Values.GetPatches()
	require.NotEmpty(t, patches)

	var setStr, removeOp bool
	for _, p := range patches {
		if p.Op == "add" && p.Path == "/some/path/to/field/str" {
			setStr = true
		}
		if p.Op == "remove" && p.Path == "/some/path/to/field" {
			removeOp = true
		}
	}
	assert.True(t, setStr, "expected Set on .str path")
	assert.True(t, removeOp, "expected Remove on .field path")
}

// The remaining cases use the typed mock since they need to inject
// behaviour the real PatchableValues cannot reproduce easily (failing
// ArrayCount, non-float GetRaw value, GetOk returning false on an
// existing path).

func TestHandlerHookValues_GetOkReturnsFalse(t *testing.T) {
	values := mock.NewOutputPatchableValuesCollectorMock(t)
	values.GetMock.When("some.path.to.field").Then(gjson.Result{Str: "str-value"})
	values.GetOkMock.When("some.path.to.field").Then(gjson.Result{Str: "str-value"}, false)
	values.ExistsMock.When("some.path.to.field.str").Then(false)
	values.SetMock.When("some.path.to.field.str", "some_string")
	values.RemoveMock.When("some.path.to.field")
	values.ArrayCountMock.When("some.path.to.field.array").Then(10, nil)

	in := &pkg.HookInput{Values: values, Logger: log.NewNop()}

	require.NoError(t, subfolder.HandlerHookValues(context.Background(), in))
}

func TestHandlerHookValues_GetRawNotFloat(t *testing.T) {
	values := mock.NewOutputPatchableValuesCollectorMock(t)
	values.GetMock.When("some.path.to.field").Then(gjson.Result{Str: "str-value"})
	values.GetOkMock.When("some.path.to.field").Then(gjson.Result{Str: "str-value"}, true)
	values.GetPatchesMock.Return(nil)
	values.GetRawMock.When("some.path.to.field.someInt").Then("not-number")
	values.ExistsMock.When("some.path.to.field.str").Then(false)
	values.SetMock.When("some.path.to.field.str", "some_string")
	values.RemoveMock.When("some.path.to.field")
	values.ArrayCountMock.When("some.path.to.field.array").Then(10, nil)

	in := &pkg.HookInput{Values: values, Logger: log.NewNop()}
	require.NoError(t, subfolder.HandlerHookValues(context.Background(), in))
}

func TestHandlerHookValues_ArrayCountReturnsError(t *testing.T) {
	wantErr := errors.New("boom")

	values := mock.NewOutputPatchableValuesCollectorMock(t)
	values.GetMock.When("some.path.to.field").Then(gjson.Result{Str: "str-value"})
	values.GetOkMock.When("some.path.to.field").Then(gjson.Result{Str: "str-value"}, true)
	values.GetPatchesMock.Return(nil)
	values.GetRawMock.When("some.path.to.field.someInt").Then(float64(1))
	values.ExistsMock.When("some.path.to.field.str").Then(false)
	values.SetMock.When("some.path.to.field.str", "some_string")
	values.RemoveMock.When("some.path.to.field")
	values.ArrayCountMock.When("some.path.to.field.array").Then(0, wantErr)

	in := &pkg.HookInput{Values: values, Logger: log.NewNop()}

	err := subfolder.HandlerHookValues(context.Background(), in)
	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
}
