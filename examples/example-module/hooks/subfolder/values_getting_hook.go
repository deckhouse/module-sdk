package hookinfolder

import (
	"context"
	"log/slog"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

// values and config values have indent interface

var _ = registry.RegisterFunc(configValues, handlerHookValues)

var configValues = &pkg.HookConfig{
	OnBeforeHelm: &pkg.OrderedConfig{Order: 1},
	Kubernetes: []pkg.KubernetesConfig{
		{
			Name:       nodeInfoSnapshotName,
			APIVersion: "v1",
			Kind:       "Node",
			JqFilter:   applyNodeJQFilter,
		},
	},
}

func handlerHookValues(ctx context.Context, input *pkg.HookInput) error {
	input.Logger.Info("hello from values hook")

	valuesGetExamples(ctx, input)

	// if field not exists - set string on it place
	if !input.Values.Exists("some.path.to.field.str") {
		input.Values.Set("some.path.to.field.str", "some_string")
	}

	// remove field
	input.Values.Remove("some.path.to.field")

	count, err := input.Values.ArrayCount("some.path.to.field.array")
	if err != nil {
		input.Logger.Error("can not extract counter from path", slog.String("error", err.Error()))

		return err
	}

	input.Logger.Info("array length counted", slog.Int("len", count))

	return nil
}

func valuesGetExamples(_ context.Context, input *pkg.HookInput) {
	// getting value from field
	fieldValStr := input.Values.Get("some.path.to.field").String()
	input.Logger.Info("value in field", slog.String("value", fieldValStr))

	// getting value from field with check existence
	fieldVal, ok := input.Values.GetOk("some.path.to.field")
	if !ok {
		input.Logger.Info("no value in field")

		return
	}

	input.Logger.Info("value in field", slog.String("value", fieldVal.String()))

	// getting all patch operations
	patchOperations := input.Values.GetPatches()
	input.Logger.Info("patches", slog.Any("patches", patchOperations))

	// returns one of these types:
	//
	//	bool, for JSON booleans
	//	float64, for JSON numbers
	//	Number, for JSON numbers
	//	string, for JSON string literals
	//	nil, for JSON null
	//	map[string]interface{}, for JSON objects
	//	[]interface{}, for JSON arrays
	//
	rawFloat := input.Values.GetRaw("some.path.to.field.someInt")
	someFloat, ok := rawFloat.(float64)
	if !ok {
		input.Logger.Info("no valid float in field")

		return
	}

	input.Logger.Info("float in field", slog.Float64("some_float", someFloat))
}
