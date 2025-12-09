package valuescheck

import (
	"context"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/app"
)

const (
	SnapshotKey = "apiservers"
)

func ValuesCheckFunc(ctx context.Context, input *pkg.HookInput) error {
	return nil
}

func main() {
	valuesCheckConfig := &app.ValuesCheckConfig{
		ProbeFunc: ValuesCheckFunc,
	}

	app.Run(app.WithValuesCheck(valuesCheckConfig))
}
