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

// HookRegistry stores registered module and application hooks.
// It is a singleton accessed via Registry().
type HookRegistry struct {
	mtx         sync.Mutex
	moduleHooks []pkg.Hook[pkg.HookConfig, *pkg.HookInput]
	appHooks    []pkg.Hook[pkg.ApplicationHookConfig, *pkg.ApplicationHookInput]
}

// Registry returns singleton instance, it is used it only in controller
func Registry() *HookRegistry {
	once.Do(func() {
		instance = &HookRegistry{
			moduleHooks: make([]pkg.Hook[pkg.HookConfig, *pkg.HookInput], 0, 1),
			appHooks:    make([]pkg.Hook[pkg.ApplicationHookConfig, *pkg.ApplicationHookInput], 0, 1),
		}
	})

	return instance
}

// ModuleHooks returns all registered module hooks.
func (h *HookRegistry) ModuleHooks() []pkg.Hook[pkg.HookConfig, *pkg.HookInput] {
	return h.moduleHooks
}

// ApplicationHooks returns all registered application hooks.
func (h *HookRegistry) ApplicationHooks() []pkg.Hook[pkg.ApplicationHookConfig, *pkg.ApplicationHookInput] {
	return h.appHooks
}

// RegisterFunc registers a hook with the global registry.
// It accepts both module and application hook configurations.
// Returns true to allow usage in var declarations: var _ = registry.RegisterFunc(...)
func RegisterFunc[C pkg.Config, T pkg.Input](config C, f pkg.HookFunc[T]) bool {
	registerHook(Registry(), config, f)
	return true
}

// registerHook validates and registers a hook with the given registry.
// It handles both pointer and value config types through type switching.
// Panics if validation fails or if OnStartup and Kubernetes bindings are mixed.
func registerHook[C pkg.Config, T pkg.Input](r *HookRegistry, cfg C, f pkg.HookFunc[T]) {
	// Phase 1: Validate OnStartup + Kubernetes conflict before extracting metadata.
	// This check must happen first to ensure proper panic ordering.
	switch c := any(cfg).(type) {
	case *pkg.HookConfig:
		if c.OnStartup != nil && len(c.Kubernetes) > 0 {
			panic(bindingsPanicMsg)
		}
	case pkg.HookConfig:
		if c.OnStartup != nil && len(c.Kubernetes) > 0 {
			panic(bindingsPanicMsg)
		}
	case *pkg.ApplicationHookConfig:
		if c.OnStartup != nil && len(c.Kubernetes) > 0 {
			panic(bindingsPanicMsg)
		}
	case pkg.ApplicationHookConfig:
		if c.OnStartup != nil && len(c.Kubernetes) > 0 {
			panic(bindingsPanicMsg)
		}
	}

	// Phase 2: Extract hook metadata from call stack (hook name and path).
	meta := extractHookMetadata()

	r.mtx.Lock()
	defer r.mtx.Unlock()

	// Phase 3: Set metadata, validate config, and register the hook.
	// Type switch is required because pkg.Config is a type constraint (union type),
	// not a regular interface - we cannot call methods on it directly.
	switch c := any(cfg).(type) {
	case *pkg.HookConfig:
		c.Metadata = meta
		if err := c.Validate(); err != nil {
			panic(fmt.Sprintf("hook validation failed for %q: %v", c.Metadata.Name, err))
		}
		hook := pkg.Hook[pkg.HookConfig, *pkg.HookInput]{
			Config:   *c,
			HookFunc: any(f).(pkg.HookFunc[*pkg.HookInput]),
		}
		r.moduleHooks = append(r.moduleHooks, hook)

	case pkg.HookConfig:
		c.Metadata = meta
		if err := c.Validate(); err != nil {
			panic(fmt.Sprintf("hook validation failed for %q: %v", c.Metadata.Name, err))
		}
		hook := pkg.Hook[pkg.HookConfig, *pkg.HookInput]{
			Config:   c,
			HookFunc: any(f).(pkg.HookFunc[*pkg.HookInput]),
		}
		r.moduleHooks = append(r.moduleHooks, hook)

	case *pkg.ApplicationHookConfig:
		c.Metadata = meta
		if err := c.Validate(); err != nil {
			panic(fmt.Sprintf("hook validation failed for %q: %v", c.Metadata.Name, err))
		}
		hook := pkg.Hook[pkg.ApplicationHookConfig, *pkg.ApplicationHookInput]{
			Config:   *c,
			HookFunc: any(f).(pkg.HookFunc[*pkg.ApplicationHookInput]),
		}
		r.appHooks = append(r.appHooks, hook)

	case pkg.ApplicationHookConfig:
		c.Metadata = meta
		if err := c.Validate(); err != nil {
			panic(fmt.Sprintf("hook validation failed for %q: %v", c.Metadata.Name, err))
		}
		hook := pkg.Hook[pkg.ApplicationHookConfig, *pkg.ApplicationHookInput]{
			Config:   c,
			HookFunc: any(f).(pkg.HookFunc[*pkg.ApplicationHookInput]),
		}
		r.appHooks = append(r.appHooks, hook)

	default:
		panic("unknown hook config type")
	}
}

// extractHookMetadata walks the call stack to extract hook name and path.
// It looks for frames matching the pattern ".../hooks/..." to determine
// the hook's location in the module structure.
// Panics if no valid hook path is found in the call stack.
func extractHookMetadata() pkg.HookMetadata {
	// Capture call stack (up to 50 frames deep)
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
		// Look for frames with ".../hooks/..." pattern
		matches := hookRe.FindStringSubmatch(frame.File)
		if matches != nil {
			// Extract hook name from path (e.g., "subfolder/my_hook" from ".../hooks/subfolder/my_hook.go")
			meta.Name = strings.TrimRight(matches[2], ".go")
			// Extract directory path
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
