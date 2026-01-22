package registry

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/deckhouse/module-sdk/pkg"
)

const bindingsPanicMsg = "OnStartup hook always has binding context without Kubernetes snapshots. To prevent logic errors, don't use OnStartup and Kubernetes bindings in the same Go hook configuration."

var (
	instance *HookRegistry
	once     sync.Once

	// /path/.../somemodule/hooks/001_ensure_crd/a/b/c/main.go
	// $1 - Hook path for values (001_ensure_crd/a/b/c/main.go)
	hookRe = regexp.MustCompile(`([^/]*)/hooks/(.*)$`)
)

type HookRegistry struct {
	mtx              sync.Mutex
	moduleHooks      []pkg.Hook[*pkg.HookInput]
	applicationHooks []pkg.Hook[*pkg.ApplicationHookInput]
}

// Registry returns singleton instance, it is used it only in controller
func Registry() *HookRegistry {
	once.Do(func() {
		instance = &HookRegistry{
			moduleHooks:      make([]pkg.Hook[*pkg.HookInput], 0, 1),
			applicationHooks: make([]pkg.Hook[*pkg.ApplicationHookInput], 0, 1),
		}
	})

	return instance
}

func (h *HookRegistry) ModuleHooks() []pkg.Hook[*pkg.HookInput] {
	return h.moduleHooks
}

func (h *HookRegistry) ApplicationHooks() []pkg.Hook[*pkg.ApplicationHookInput] {
	return h.applicationHooks
}

func RegisterFunc[T pkg.Input](config *pkg.HookConfig, f pkg.HookFunc[T]) bool {
	registerHook(Registry(), config, f)
	return true
}

func registerHook[T pkg.Input](r *HookRegistry, cfg *pkg.HookConfig, f pkg.HookFunc[T]) {
	if cfg.OnStartup != nil && len(cfg.Kubernetes) > 0 {
		panic(bindingsPanicMsg)
	}

	if cfg.Metadata.Name == "" {
		cfg.Metadata = extractHookMetadata()
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	hook := pkg.Hook[T]{Config: cfg, HookFunc: f}

	switch any(hook).(type) {
	case pkg.Hook[*pkg.HookInput]:
		cfg.HookType = pkg.HookTypeModule
		r.moduleHooks = append(r.moduleHooks, any(hook).(pkg.Hook[*pkg.HookInput]))
	case pkg.Hook[*pkg.ApplicationHookInput]:
		cfg.HookType = pkg.HookTypeApplication
		r.applicationHooks = append(r.applicationHooks, any(hook).(pkg.Hook[*pkg.ApplicationHookInput]))
	default:
		panic("unknown hook input type")
	}

	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("hook validation failed for %q: %v", cfg.Metadata.Name, err))
	}
}

func extractHookMetadata() pkg.HookMetadata {
	pc := make([]uintptr, 50)
	n := runtime.Callers(0, pc)
	if n == 0 {
		panic("runtime.Callers is empty")
	}
	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	meta := pkg.HookMetadata{}
	for {
		frame, more := frames.Next()
		matches := hookRe.FindStringSubmatch(frame.File)
		if matches != nil {
			meta.Name = strings.TrimRight(matches[2], ".go")
			lastSlashIdx := strings.LastIndex(matches[0], "/")
			meta.Path = matches[0][:lastSlashIdx+1]
		}
		if !more {
			break
		}
	}

	if len(meta.Name) == 0 {
		panic("cannot extract metadata from GoHook")
	}

	return meta
}
