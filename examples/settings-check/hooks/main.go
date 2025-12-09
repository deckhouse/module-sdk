package main

import (
	"context"

	"github.com/deckhouse/module-sdk/pkg/app"
	"github.com/deckhouse/module-sdk/pkg/settingscheck"
)

func settingsCheck(_ context.Context, input settingscheck.Input) error {
	if !input.Settings.Get(".enabled").Bool() {
		return &settingscheck.Warning{Message: "settings disabled"}
	}

	return nil
}

func main() {
	app.Run(app.WithSettingsCheck(settingsCheck))
}
