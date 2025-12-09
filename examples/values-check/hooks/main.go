package main

import (
	"context"

	"github.com/deckhouse/module-sdk/pkg/app"
	settingscheck "github.com/deckhouse/module-sdk/pkg/settings-check"
)

const (
	SnapshotKey = "apiservers"
)

func settingsCheckFunc(ctx context.Context, input *settingscheck.SettingsCheckHookInput) settingscheck.SettingsCheckHookResult {
	res := settingscheck.SettingsCheckHookResult{
		Allow: true,
	}

	if true {
		res.Allow = false
		res.Message = "this is a test warning"
	}

	return res
}

func main() {
	app.Run(app.WithSettingsCheck(settingsCheckFunc))
}
