package hooks

import (
	"context"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

var _ = registry.RegisterFunc(&pkg.HookConfig{}, second)

func second(_ context.Context, _ *pkg.HookInput) error {
	return nil
}
