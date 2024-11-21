package hooks

import (
	pkg "github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

var _ = registry.RegisterFunc(&pkg.HookConfig{}, first)

func first(_ *pkg.HookInput) error {
	return nil
}
