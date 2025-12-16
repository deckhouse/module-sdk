package main

import (
	"context"

	"github.com/deckhouse/module-sdk/pkg/app"
	"github.com/deckhouse/module-sdk/pkg/settingscheck"
)

func check(_ context.Context, input settingscheck.Input) settingscheck.Result {
	replicas := input.Settings.Get("replicas").Int()
	if replicas == 0 {
		return settingscheck.Reject("replicas cannot be 0")
	}

	if replicas == 2 {
		return settingscheck.Warn("replicas cannot be greater than 2")
	}

	if replicas > 3 {
		return settingscheck.Reject("replicas cannot be greater than 3")
	}

	return settingscheck.Allow()
}

func main() {
	app.Run(app.WithSettingsCheck(check))
}
