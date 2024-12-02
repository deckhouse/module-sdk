package internal

import (
	bindingcontext "github.com/deckhouse/module-sdk/internal/binding-context"
	"github.com/deckhouse/module-sdk/pkg"
)

type HookRequest interface {
	GetValues() (map[string]any, error)
	GetConfigValues() (map[string]any, error)
	GetBindingContexts() ([]bindingcontext.BindingContext, error)
	GetDependencyContainer() pkg.DependencyContainer
}
