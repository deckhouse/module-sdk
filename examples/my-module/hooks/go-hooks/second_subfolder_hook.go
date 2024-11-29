package hookinsubfolder

import (
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

var configSecond = &pkg.HookConfig{
	OnBeforeHelm: &pkg.OrderedConfig{Order: 1},
}

var _ = registry.RegisterFunc(configSecond, handlerCRDSecond)

func handlerCRDSecond(input *pkg.HookInput) error {
	input.Logger.Info("hello from second subfolder hook")

	return nil
}
