package hookinfolder

import (
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

var _ = registry.RegisterFunc(configSecond, handlerHookSecond)

var configSecond = &pkg.HookConfig{}

func handlerHookSecond(input *pkg.HookInput) error {
	input.Logger.Info("hello from second root hook")

	return nil
}
