package registry

import (
	"fmt"
	"regexp"
	"runtime"
	"sync"

	"github.com/deckhouse/deckhouse/pkg/log"
	gohook "github.com/deckhouse/module-sdk/pkg/hook"
)

const bindingsPanicMsg = "OnStartup hook always has binding context without Kubernetes snapshots. To prevent logic errors, don't use OnStartup and Kubernetes bindings in the same Go hook configuration."

// /path/.../somemodule/hooks/go-hooks/001_ensure_crd/a/b/c/main.go
// $1 - Hook path for values (/001_ensure_crd/a/b/c/main.go)
// $2 - Hook name for identification (001_ensure_crd)
var hookRe = regexp.MustCompile(`/hooks/go-hooks/(([^/]+)(/([^/]+/)*([^/]+)))$`)

// simple_module/hooks/go-hooks/002-hook-two/level1/sublevel

var RegisterFunc = func(hook *gohook.GoHook) bool {
	Registry().Add(hook)
	return true
}

type HookRegistry struct {
	m      sync.Mutex
	hooks  []*gohook.GoHook
	logger *log.Logger
}

var (
	instance *HookRegistry
	once     sync.Once
)

func Registry() *HookRegistry {
	once.Do(func() {
		logger := log.NewLogger(log.Options{})
		log.SetDefault(logger)

		instance = &HookRegistry{
			hooks:  make([]*gohook.GoHook, 0, 1),
			logger: logger,
		}
	})
	return instance
}

func (h *HookRegistry) SetLogLevel(lvl log.Level) {
	h.logger.SetLevel(lvl)
}

// Hooks returns all (module and global) hooks
// Deprecated: method exists for backward compatibility, use GetGlobalHooks or GetModuleHooks instead
func (h *HookRegistry) Hooks() []*gohook.GoHook {
	return h.hooks
}

func (h *HookRegistry) Add(hook *gohook.GoHook) {
	config := hook.GetConfig()
	if config.OnStartup != 0 && len(config.Kubernetes) > 0 {
		panic(bindingsPanicMsg)
	}

	pc := make([]uintptr, 50)
	n := runtime.Callers(0, pc)
	if n == 0 {
		panic("runtime.Callers is empty")
	}
	pc = pc[:n] // pass only valid pcs to runtime.CallersFrames
	frames := runtime.CallersFrames(pc)

	for {
		frame, more := frames.Next()

		matches := hookRe.FindStringSubmatch(frame.File)
		fmt.Println(frame.File)
		if matches != nil {
			hook.Name = matches[2]
			hook.Path = matches[1]
		}

		if !more {
			break
		}
	}

	if len(hook.Name) == 0 {
		panic("cannot extract metadata from GoHook")
	}

	hook.SetLogger(h.logger.Named(hook.Name))

	h.m.Lock()
	defer h.m.Unlock()

	h.hooks = append(h.hooks, hook)
}
