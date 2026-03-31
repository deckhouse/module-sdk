package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/internal/common-hooks/readiness"
	execregistry "github.com/deckhouse/module-sdk/internal/executor/registry"
	"github.com/deckhouse/module-sdk/internal/transport/file"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/dependency"
	gohook "github.com/deckhouse/module-sdk/pkg/hook"
	hookregistry "github.com/deckhouse/module-sdk/pkg/registry"
	"github.com/deckhouse/module-sdk/pkg/settingscheck"
	"github.com/deckhouse/module-sdk/pkg/utils/ptr"
)

type HookController struct {
	registry *execregistry.Registry
	fConfig  *file.Config

	settingsCheck settingscheck.Check

	dc     pkg.DependencyContainer
	logger *log.Logger
}

type HookSender interface {
	SendMetrics() error
	SendPatches() error
	SendValues() error
	SendConfigValues() error
}

func NewHookController(cfg *Config, logger *log.Logger) *HookController {
	reg := execregistry.NewRegistry(logger)
	reg.RegisterModuleHooks(hookregistry.Registry().ModuleHooks()...)
	reg.RegisterAppHooks(hookregistry.Registry().ApplicationHooks()...)

	if cfg.ReadinessConfig != nil {
		addReadinessHook(reg, cfg.ReadinessConfig)
	}

	return &HookController{
		registry:      reg,
		settingsCheck: cfg.SettingsCheck,
		dc:            dependency.NewDependencyContainer(),
		fConfig:       cfg.GetFileConfig(),
		logger:        logger,
	}
}

func addReadinessHook(reg *execregistry.Registry, cfg *ReadinessConfig) {
	readinessConfig := &readiness.ReadinessHookConfig{
		ModuleName:        cfg.ModuleName,
		IntervalInSeconds: cfg.IntervalInSeconds,
		ProbeFunc:         cfg.ProbeFunc,
	}

	config, f := readiness.NewReadinessHookEM(readinessConfig)
	config.Metadata.Name = "readiness"
	config.Metadata.Path = "common-hooks/readiness"

	reg.SetReadinessHook(pkg.Hook[pkg.HookConfig, *pkg.HookInput]{Config: *config, HookFunc: f})
}

func (c *HookController) ListHooksMeta() []pkg.HookMetadata {
	hooks := c.registry.Executors()

	hooksmetas := make([]pkg.HookMetadata, 0, len(hooks))
	for _, hook := range hooks {
		hooksmetas = append(hooksmetas, hook.Config().GetMetadata())
	}

	return hooksmetas
}

// TODO: fix typo, didn't fix now to not break public API
var ErrHookIndexIsNotExists = errors.New("hook index does not exist")

func (c *HookController) RunHook(ctx context.Context, idx int) error {
	hooks := c.registry.Executors()

	if len(hooks) <= idx {
		return ErrHookIndexIsNotExists
	}

	hook := hooks[idx]

	transport := file.NewTransport(c.fConfig, hook.Config().GetMetadata().Name, c.dc, c.logger.Named("file-transport"))

	hookRes, err := hook.Execute(ctx, transport.NewRequest())
	if err != nil {
		outputError := &gohook.Error{Message: "execute: " + err.Error()}

		buf := bytes.NewBuffer([]byte{})
		err := json.NewEncoder(buf).Encode(outputError)
		if err != nil {
			return fmt.Errorf("encode error: %w", err)
		}

		fmt.Fprintln(os.Stderr, buf.String())
		os.Exit(1)
	}

	err = transport.NewResponse().Send(hookRes)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}

var ErrReadinessHookDoesNotExists = errors.New("readiness hook does not exists")

func (c *HookController) RunReadiness(ctx context.Context) error {
	hook := c.registry.Readiness()

	if hook == nil {
		return ErrReadinessHookDoesNotExists
	}

	transport := file.NewTransport(c.fConfig, hook.Config().GetMetadata().Name, c.dc, c.logger.Named("file-transport"))

	hookRes, err := hook.Execute(ctx, transport.NewRequest())
	if err != nil {
		outputError := &gohook.Error{Message: "execute: " + err.Error()}

		buf := bytes.NewBuffer([]byte{})
		err := json.NewEncoder(buf).Encode(outputError)
		if err != nil {
			return fmt.Errorf("encode error: %w", err)
		}

		fmt.Fprintln(os.Stderr, buf.String())
		os.Exit(1)
	}

	err = transport.NewResponse().Send(hookRes)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}

func (c *HookController) CheckSettings(ctx context.Context) error {
	res := settingscheck.Execute(ctx, c.settingsCheck, c.dc, c.logger)

	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(res); err != nil {
		return fmt.Errorf("encode error: %w", err)
	}

	fmt.Fprintln(os.Stderr, buf.String())
	os.Exit(1)

	return nil
}

var ErrNoHooksRegistered = errors.New("no hooks registered")

func (c *HookController) PrintHookConfigs() error {
	if len(c.registry.Executors()) == 0 && c.settingsCheck == nil && c.registry.Readiness() == nil {
		return ErrNoHooksRegistered
	}

	configs := make([]gohook.HookConfig, 0, 1)

	for _, hook := range c.registry.Executors() {
		hookConfig := remapHookConfigToGohook(hook.Config())
		configs = append(configs, *hookConfig)
	}

	cfg := &gohook.BatchHookConfig{
		Version: gohook.BatchHookConfigV1,
		Hooks:   configs,
	}

	if c.registry.Readiness() != nil {
		readinessConfig := remapHookConfigToGohook(c.registry.Readiness().Config())
		cfg.Readiness = readinessConfig
	}

	if c.settingsCheck != nil {
		cfg.HasSettingsCheck = true
	}

	buf := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(buf).Encode(cfg)
	if err != nil {
		return fmt.Errorf("json encode: %w", err)
	}

	fmt.Print(buf.String())

	return nil
}

func (c *HookController) WriteHookConfigsInFile() error {
	if len(c.registry.Executors()) == 0 && c.settingsCheck == nil && c.registry.Readiness() == nil {
		return ErrNoHooksRegistered
	}

	var configsFileName = c.fConfig.HookConfigPath

	if c.fConfig.CreateFilesByYourself {
		dir := filepath.Dir(configsFileName)

		err := os.MkdirAll(dir, 0744)
		if err != nil {
			return fmt.Errorf("mkdir all: %w", err)
		}
	}

	f, err := os.OpenFile(configsFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	configs := make([]gohook.HookConfig, 0, 1)

	for _, hook := range c.registry.Executors() {
		hookConfig := remapHookConfigToGohook(hook.Config())
		configs = append(configs, *hookConfig)
	}

	cfg := &gohook.BatchHookConfig{
		Version: "v1",
		Hooks:   configs,
	}

	if c.registry.Readiness() != nil {
		readinessConfig := remapHookConfigToGohook(c.registry.Readiness().Config())
		cfg.Readiness = readinessConfig
	}

	err = json.NewEncoder(f).Encode(cfg)
	if err != nil {
		return fmt.Errorf("json marshall: %w", err)
	}

	return nil
}

// remapHookConfigToGohook converts HookConfigLike to gohook.HookConfig for shell-operator.
func remapHookConfigToGohook(cfg pkg.HookConfigInterface) *gohook.HookConfig {
	out := &gohook.HookConfig{
		ConfigVersion: "v1",
		Metadata:      gohook.GoHookMetadata(cfg.GetMetadata()),
	}
	if c, ok := cfg.AsHookConfig(); ok {
		remapModuleHookConfig(c, out)
	} else if c, ok := cfg.AsApplicationHookConfig(); ok {
		remapApplicationHookConfig(c, out)
	}
	return out
}

func remapModuleHookConfig(cfg *pkg.HookConfig, out *gohook.HookConfig) {
	for _, scfg := range cfg.Schedule {
		out.Schedule = append(out.Schedule, gohook.ScheduleConfig{
			Name:    scfg.Name,
			Crontab: scfg.Crontab,
			Queue:   cfg.Queue,
		})
	}

	for i := range cfg.Kubernetes {
		k := &cfg.Kubernetes[i]
		out.Kubernetes = append(out.Kubernetes, convertKubernetesConfig(k, cfg.Queue))
	}

	if cfg.OnStartup != nil {
		out.OnStartup = ptr.To(cfg.OnStartup.Order)
	}
	if cfg.OnBeforeHelm != nil {
		out.OnBeforeHelm = ptr.To(cfg.OnBeforeHelm.Order)
	}
	if cfg.OnAfterHelm != nil {
		out.OnAfterHelm = ptr.To(cfg.OnAfterHelm.Order)
	}
	if cfg.OnAfterDeleteHelm != nil {
		out.OnAfterDeleteHelm = ptr.To(cfg.OnAfterDeleteHelm.Order)
	}
}

func remapApplicationHookConfig(cfg *pkg.ApplicationHookConfig, out *gohook.HookConfig) {
	for _, scfg := range cfg.Schedule {
		out.Schedule = append(out.Schedule, gohook.ScheduleConfig{
			Name:    scfg.Name,
			Crontab: scfg.Crontab,
			Queue:   cfg.Queue,
		})
	}

	for i := range cfg.Kubernetes {
		k := &cfg.Kubernetes[i]
		out.Kubernetes = append(out.Kubernetes, convertAppKubernetesConfig(k, cfg.Queue))
	}

	if cfg.OnStartup != nil {
		out.OnStartup = ptr.To(cfg.OnStartup.Order)
	}
	if cfg.OnBeforeHelm != nil {
		out.OnBeforeHelm = ptr.To(cfg.OnBeforeHelm.Order)
	}
	if cfg.OnAfterHelm != nil {
		out.OnAfterHelm = ptr.To(cfg.OnAfterHelm.Order)
	}
	if cfg.OnAfterDeleteHelm != nil {
		out.OnAfterDeleteHelm = ptr.To(cfg.OnAfterDeleteHelm.Order)
	}
}

func convertKubernetesConfig(k *pkg.KubernetesConfig, queue string) gohook.KubernetesConfig {
	cfg := gohook.KubernetesConfig{
		APIVersion:                   k.APIVersion,
		Kind:                         k.Kind,
		Name:                         k.Name,
		LabelSelector:                k.LabelSelector,
		ExecuteHookOnEvents:          k.ExecuteHookOnEvents,
		ExecuteHookOnSynchronization: k.ExecuteHookOnSynchronization,
		WaitForSynchronization:       k.WaitForSynchronization,
		KeepFullObjectsInMemory:      ptr.To(k.JqFilter == ""),
		JqFilter:                     k.JqFilter,
		AllowFailure:                 k.AllowFailure,
		ResynchronizationPeriod:      k.ResynchronizationPeriod,
		Queue:                        queue,
	}

	if k.NameSelector != nil {
		cfg.NameSelector = &gohook.NameSelector{MatchNames: k.NameSelector.MatchNames}
	}
	if k.NamespaceSelector != nil {
		cfg.NamespaceSelector = &gohook.NamespaceSelector{
			NameSelector:  &gohook.NameSelector{MatchNames: k.NamespaceSelector.NameSelector.MatchNames},
			LabelSelector: k.NamespaceSelector.LabelSelector,
		}
	}
	if k.FieldSelector != nil {
		fs := &gohook.FieldSelector{
			MatchExpressions: make([]gohook.FieldSelectorRequirement, 0, len(k.FieldSelector.MatchExpressions)),
		}
		for _, expr := range k.FieldSelector.MatchExpressions {
			fs.MatchExpressions = append(fs.MatchExpressions, gohook.FieldSelectorRequirement(expr))
		}
		cfg.FieldSelector = fs
	}

	return cfg
}

func convertAppKubernetesConfig(k *pkg.ApplicationKubernetesConfig, queue string) gohook.KubernetesConfig {
	cfg := gohook.KubernetesConfig{
		APIVersion:                   k.APIVersion,
		Kind:                         k.Kind,
		Name:                         k.Name,
		LabelSelector:                k.LabelSelector,
		ExecuteHookOnEvents:          k.ExecuteHookOnEvents,
		ExecuteHookOnSynchronization: k.ExecuteHookOnSynchronization,
		WaitForSynchronization:       k.WaitForSynchronization,
		KeepFullObjectsInMemory:      ptr.To(k.JqFilter == ""),
		JqFilter:                     k.JqFilter,
		AllowFailure:                 k.AllowFailure,
		ResynchronizationPeriod:      k.ResynchronizationPeriod,
		Queue:                        queue,
	}

	if k.NameSelector != nil {
		cfg.NameSelector = &gohook.NameSelector{MatchNames: k.NameSelector.MatchNames}
	}
	if k.FieldSelector != nil {
		fs := &gohook.FieldSelector{
			MatchExpressions: make([]gohook.FieldSelectorRequirement, 0, len(k.FieldSelector.MatchExpressions)),
		}
		for _, expr := range k.FieldSelector.MatchExpressions {
			fs.MatchExpressions = append(fs.MatchExpressions, gohook.FieldSelectorRequirement(expr))
		}
		cfg.FieldSelector = fs
	}

	return cfg
}
