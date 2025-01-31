package registry

import (
	"github.com/deckhouse/deckhouse/pkg/log"
	gohook "github.com/deckhouse/module-sdk/internal/hook"
	"github.com/deckhouse/module-sdk/pkg"
)

type HookRegistry struct {
	hooks  []*gohook.Hook
	logger *log.Logger
}

func NewHookRegistry(logger *log.Logger) *HookRegistry {
	return &HookRegistry{
		hooks:  make([]*gohook.Hook, 0, 1),
		logger: logger,
	}
}

// Hooks returns all hooks
func (h *HookRegistry) Hooks() []*gohook.Hook {
	return h.hooks
}

func (h *HookRegistry) Add(hooks ...*pkg.Hook) {
	for _, hook := range hooks {
		newHook := gohook.NewHook(hook.Config, hook.ReconcileFunc)
		newHook.SetLogger(h.logger.Named(newHook.GetName()))

		h.hooks = append(h.hooks, newHook)
	}
}
