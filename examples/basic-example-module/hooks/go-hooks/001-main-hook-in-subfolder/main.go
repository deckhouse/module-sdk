package mainhookinsubfolder

import (
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

var _ = registry.RegisterFunc(config, handlerHook)

var config = &pkg.HookConfig{}

func handlerHook(input *pkg.HookInput) error {
	input.Logger.Info("hello from main hook in subfolder")

	return nil
}
