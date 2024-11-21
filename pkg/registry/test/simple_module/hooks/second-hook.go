package hooks

import (
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

var _ = registry.RegisterFunc(&pkg.HookConfig{}, second)

func second(_ *pkg.HookInput) error {
	return nil
}
