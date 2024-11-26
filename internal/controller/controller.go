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
	"github.com/deckhouse/module-sdk/pkg/utils/ptr"
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

var ErrNoHooksRegistered = errors.New("no hooks registered")

func (c *HookController) WriteHookConfigsInFile() error {
	if len(c.registry.Hooks()) == 0 {
		return ErrNoHooksRegistered
	}

	const configsPath = "configs.json"

	f, err := os.OpenFile(configsPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	configs := make([]*hook.HookConfig, 0, 1)

	for _, hook := range c.registry.Hooks() {
		configs = append(configs, remapHookConfigToHookConfig(hook.GetConfig()))
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
		ConfigVersion: cfg.ConfigVersion,
		Metadata:      hook.GoHookMetadata(cfg.Metadata),
	}

	for _, scfg := range cfg.Schedule {
		newHookConfig.Schedule = append(newHookConfig.Schedule, hook.ScheduleConfig{
			Name:    scfg.Name,
			Crontab: scfg.Crontab,
		})
	}

	for _, shcfg := range cfg.Kubernetes {
		newShCfg := hook.KubernetesConfig{
			APIVersion:                   shcfg.APIVersion,
			Kind:                         shcfg.Kind,
			NameSelector:                 (*hook.NameSelector)(shcfg.NameSelector),
			LabelSelector:                shcfg.LabelSelector,
			ExecuteHookOnEvents:          shcfg.ExecuteHookOnEvents,
			ExecuteHookOnSynchronization: shcfg.ExecuteHookOnSynchronization,
			WaitForSynchronization:       shcfg.WaitForSynchronization,
			KeepFullObjectsInMemory:      ptr.To(false),
			JqFilter:                     shcfg.JqFilter,
			AllowFailure:                 shcfg.AllowFailure,
			ResynchronizationPeriod:      shcfg.ResynchronizationPeriod,
			IncludeSnapshotsFrom:         shcfg.IncludeSnapshotsFrom,
			Queue:                        shcfg.Queue,
			Group:                        shcfg.Group,
		}

		if shcfg.NamespaceSelector != nil {
			newShCfg.NamespaceSelector = &hook.NamespaceSelector{
				NameSelector:  (*hook.NameSelector)(shcfg.NameSelector),
				LabelSelector: shcfg.LabelSelector,
			}
		}

		if shcfg.FieldSelector != nil {
			fs := &hook.FieldSelector{
				MatchExpressions: make([]hook.FieldSelectorRequirement, 0, len(shcfg.FieldSelector.MatchExpressions)),
			}

			for _, expr := range shcfg.FieldSelector.MatchExpressions {
				fs.MatchExpressions = append(fs.MatchExpressions, hook.FieldSelectorRequirement(expr))
			}

			newShCfg.FieldSelector = fs
		}

		newHookConfig.Kubernetes = append(newHookConfig.Kubernetes, newShCfg)
	}

	if cfg.OnStartup != nil {
		newHookConfig.OnStartup = ptr.To(cfg.OnStartup.Order)
	}

	if cfg.OnBeforeHelm != nil {
		newHookConfig.OnBeforeHelm = ptr.To(cfg.OnBeforeHelm.Order)
	}

	if cfg.OnAfterHelm != nil {
		newHookConfig.OnAfterHelm = ptr.To(cfg.OnAfterHelm.Order)
	}

	if cfg.OnAfterDeleteHelm != nil {
		newHookConfig.OnAfterDeleteHelm = ptr.To(cfg.OnAfterDeleteHelm.Order)
	}

	return newHookConfig
}
