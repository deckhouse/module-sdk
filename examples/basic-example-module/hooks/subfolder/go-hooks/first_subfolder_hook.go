package hookinsubfolder

import (
	"context"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

var _ = registry.RegisterFunc(config, handlerHook)

var config = &pkg.HookConfig{}

func handlerHook(_ context.Context, input *pkg.HookInput) error {
	input.Logger.Info("hello from first subfolder hook")

	return nil
}
