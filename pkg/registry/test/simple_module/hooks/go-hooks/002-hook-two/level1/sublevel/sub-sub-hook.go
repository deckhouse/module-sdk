package sublevel

import (
	gohook "github.com/deckhouse/module-sdk/pkg/hook"
	gohookcfg "github.com/deckhouse/module-sdk/pkg/hook/config"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

var _ = registry.RegisterFunc(gohook.NewGoHook(&gohookcfg.HookConfig{}, main))

func main(_ *gohook.HookInput) error {
	return nil
}
