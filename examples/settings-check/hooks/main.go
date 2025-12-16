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

	var warnings []string
	if replicas == 2 {
		warnings = append(warnings, "replicas cannot be greater than 3")
	}

	if replicas > 3 {
		return settingscheck.Reject("replicas cannot be greater than 3")
	}

	return settingscheck.Allow(warnings...)
}

func main() {
	app.Run(app.WithSettingsCheck(check))
}
