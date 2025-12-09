package main

import (
	"context"

	"github.com/deckhouse/module-sdk/pkg/app"
	settingscheck "github.com/deckhouse/module-sdk/pkg/settings-check"
)

func settingsCheckFunc(_ context.Context, input *settingscheck.SettingsCheckHookInput) settingscheck.SettingsCheckHookResult {
	res := settingscheck.SettingsCheckHookResult{
		Allow: true,
	}

	test := input.Values.Get("global.test").Bool()

	if !test {
		res.Allow = false
		res.Message = "this is a test warning"
	}

	return res
}

func main() {
	app.Run(app.WithSettingsCheck(settingsCheckFunc))
}
