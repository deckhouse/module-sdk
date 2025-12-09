package main

import (
	"context"
	"fmt"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/app"
)

const (
	SnapshotKey = "apiservers"
)

func SettingsCheckFunc(ctx context.Context, input *pkg.HookInput) error {
	fmt.Println("settings check")
	return nil
}

func main() {
	app.Run(app.WithSettingsCheck(SettingsCheckFunc))
}
