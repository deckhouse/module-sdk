package registry

import (
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/deckhouse/module-sdk/pkg"
)

const bindingsPanicMsg = "OnStartup hook always has binding context without Kubernetes snapshots. To prevent logic errors, don't use OnStartup and Kubernetes bindings in the same Go hook configuration."

// /path/.../somemodule/hooks/001_ensure_crd/a/b/c/main.go
// $1 - Hook path for values (001_ensure_crd/a/b/c/main.go)
var hookRe = regexp.MustCompile(`([^/]*)/hooks/(.*)$`)

var RegisterFunc = func(config *pkg.HookConfig, f pkg.ReconcileFunc) bool {
	Registry().Add(&pkg.Hook{Config: config, ReconcileFunc: f})
	return true
}

type HookRegistry struct {
	m     sync.Mutex
	hooks []*pkg.Hook
}

var (
	instance *HookRegistry
	once     sync.Once
)

// use it only in controller
func Registry() *HookRegistry {
	once.Do(func() {
		instance = &HookRegistry{
			hooks: make([]*pkg.Hook, 0, 1),
		}
	})
	return instance
}

// Hooks returns all hooks
func (h *HookRegistry) Hooks() []*pkg.Hook {
	return h.hooks
}

func (h *HookRegistry) Add(hook *pkg.Hook) {
	config := hook.Config
	if config.OnStartup != nil && len(config.Kubernetes) > 0 {
		panic(bindingsPanicMsg)
	}

	pc := make([]uintptr, 50)
	n := runtime.Callers(0, pc)
	if n == 0 {
		panic("runtime.Callers is empty")
	}
	pc = pc[:n] // pass only valid pcs to runtime.CallersFrames
	frames := runtime.CallersFrames(pc)

	meta := pkg.GoHookMetadata{}

	for {
		frame, more := frames.Next()

		matches := hookRe.FindStringSubmatch(frame.File)
		if matches != nil {
			meta.Name = strings.TrimRight(matches[2], ".go")

			lastSlashIdx := strings.LastIndex(matches[0], "/")
			// trim with last slash

			meta.Path = matches[0][:lastSlashIdx+1]
		}

		if !more {
			break
		}
	}

	hook.Config.Metadata = meta

	if len(hook.Config.Metadata.Name) == 0 {
		panic("cannot extract metadata from GoHook")
	}

	h.m.Lock()
	defer h.m.Unlock()

	h.hooks = append(h.hooks, hook)
}
