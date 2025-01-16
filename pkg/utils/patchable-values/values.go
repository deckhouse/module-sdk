package patchablevalues

import (
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/tidwall/gjson"
)

func getFirstDefined(values pkg.PatchableValuesCollector, keys ...string) (gjson.Result, bool) {
	for i := range keys {
		v, ok := values.GetOk(keys[i])
		if ok {
			return v, ok
		}
	}

	return gjson.Result{}, false
}

func GetValuesFirstDefined(input *pkg.HookInput, keys ...string) (gjson.Result, bool) {
	return getFirstDefined(input.Values, keys...)
}

func GetConfigValuesFirstDefined(input *pkg.HookInput, keys ...string) (gjson.Result, bool) {
	return getFirstDefined(input.ConfigValues, keys...)
}

func GetHTTPSMode(input *pkg.HookInput, moduleName string) string {
	var (
		modulePath = moduleName + ".https.mode"
		globalPath = "global.modules.https.mode"
	)

	v, ok := GetValuesFirstDefined(input, modulePath, globalPath)
	if ok {
		return v.String()
	}

	panic("https mode is not defined")
}
