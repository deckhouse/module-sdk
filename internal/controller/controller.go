package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/deckhouse/module-sdk/internal/hook"
	"github.com/deckhouse/module-sdk/internal/registry"
	"github.com/deckhouse/module-sdk/internal/transport/file"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/dependency"
	outerRegistry "github.com/deckhouse/module-sdk/pkg/registry"
)

type HookController struct {
	registry *registry.HookRegistry
	dc       pkg.DependencyContainer

	fConfig file.Config

	logger *log.Logger
}

type HookSender interface {
	SendMetrics() error
	SendPatches() error
	SendValues() error
	SendConfigValues() error
}

func NewHookController(fConfig file.Config, logger *log.Logger) *HookController {
	reg := registry.NewHookRegistry(logger)
	reg.Add(outerRegistry.Registry().Hooks()...)

	return &HookController{
		registry: reg,
		dc:       dependency.NewDependencyContainer(),
		fConfig:  fConfig,
		logger:   logger,
	}
}

func (c *HookController) ListHooksMeta() []pkg.GoHookMetadata {
	hooks := c.registry.Hooks()

	hooksmetas := make([]pkg.GoHookMetadata, 0, len(hooks))
	for _, hook := range hooks {
		hooksmetas = append(hooksmetas, hook.GetConfig().Metadata)
	}

	return hooksmetas
}

var ErrHookIndexIsNotExists = errors.New("hook index is not exists")

func (c *HookController) RunHook(idx int) error {
	hooks := c.registry.Hooks()

	if len(hooks) <= idx {
		return ErrHookIndexIsNotExists
	}

	hook := hooks[idx]

	transport := file.NewTransport(c.fConfig, hook.GetName(), c.dc, c.logger.Named("file-transport"))

	hookRes, err := hook.Execute(transport.NewRequest())
	if err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	err = transport.NewResponse().Send(hookRes)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}

func (c *HookController) WriteHookConfigsInFile() error {
	const configsPath = "configs.json"

	f, err := os.OpenFile(configsPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	configs := make(map[string]*hook.HookConfig, 1)

	for _, hook := range c.registry.Hooks() {
		configs[hook.GetName()] = remapHookConfigToHookConfig(hook.GetConfig())
	}

	err = json.NewEncoder(f).Encode(configs)
	if err != nil {
		return fmt.Errorf("json marshall: %w", err)
	}

	return nil
}

func remapHookConfigToHookConfig(cfg *pkg.HookConfig) *hook.HookConfig {
	// TODO: complete remap
	newHookConfig := &hook.HookConfig{
		Metadata:      hook.GoHookMetadata(cfg.Metadata),
		ConfigVersion: cfg.ConfigVersion,
	}

	return newHookConfig
}
