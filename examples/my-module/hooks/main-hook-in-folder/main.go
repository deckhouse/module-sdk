package hookinfolder

import (
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

var config = &pkg.HookConfig{
	OnBeforeHelm: &pkg.OrderedConfig{Order: 1},
}

var _ = registry.RegisterFunc(config, handlerCRD)

func handlerCRD(input *pkg.HookInput) error {
	input.Logger.Info("hello from main hook in folder")

	return nil
}
