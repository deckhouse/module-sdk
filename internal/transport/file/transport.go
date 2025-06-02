package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/deckhouse/deckhouse/pkg/log"

	bindingcontext "github.com/deckhouse/module-sdk/internal/binding-context"
	"github.com/deckhouse/module-sdk/internal/hook"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/utils"
)

type Config struct {
	// input
	BindingContextPath string
	ValuesPath         string
	ConfigValuesPath   string

	// output
	MetricsPath          string
	KubernetesPath       string
	ValuesJSONPath       string
	ConfigValuesJSONPath string

	HookConfigPath string

	CreateFilesByYourself bool
}

type Transport struct {
	hookName string

	// input
	BindingContextPath string
	ValuesPath         string
	ConfigValuesPath   string

	// output
	MetricsPath          string
	KubernetesPath       string
	ValuesJSONPath       string
	ConfigValuesJSONPath string

	dc pkg.DependencyContainer

	CreateFilesByYourself bool

	logger *log.Logger
}

func NewTransport(cfg *Config, hookName string, dc pkg.DependencyContainer, logger *log.Logger) *Transport {
	if cfg == nil {
		panic("transport config is nil")
	}

	return &Transport{
		hookName: hookName,

		BindingContextPath: cfg.BindingContextPath,
		ValuesPath:         cfg.ValuesPath,
		ConfigValuesPath:   cfg.ConfigValuesPath,

		MetricsPath:          cfg.MetricsPath,
		KubernetesPath:       cfg.KubernetesPath,
		ValuesJSONPath:       cfg.ValuesJSONPath,
		ConfigValuesJSONPath: cfg.ConfigValuesJSONPath,

		dc: dc,

		CreateFilesByYourself: cfg.CreateFilesByYourself,

		logger: logger,
	}
}

func (t *Transport) NewRequest() *Request {
	return &Request{
		hookName: t.hookName,

		BindingContextPath: t.BindingContextPath,
		ValuesPath:         t.ValuesPath,
		ConfigValuesPath:   t.ConfigValuesPath,

		dc: t.dc,

		logger: t.logger,
	}
}

type Request struct {
	hookName string

	BindingContextPath string
	ValuesPath         string
	ConfigValuesPath   string

	dc pkg.DependencyContainer

	logger *log.Logger
}

func (r *Request) GetValues() (map[string]any, error) {
	values, err := r.loadValuesFromFile(r.ValuesPath)
	if err != nil {
		return nil, fmt.Errorf("load values from file: %w", err)
	}

	return values, nil
}

func (r *Request) GetConfigValues() (map[string]any, error) {
	values, err := r.loadValuesFromFile(r.ConfigValuesPath)
	if err != nil {
		return nil, fmt.Errorf("load values from file: %w", err)
	}

	return values, nil
}

func (r *Request) GetBindingContexts() ([]bindingcontext.BindingContext, error) {
	contextsContent, err := os.Open(r.BindingContextPath)
	defer func() {
		err := contextsContent.Close()
		if err != nil {
			r.logger.Error("binding contexts file close", slog.String("error", err.Error()))
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("open binding context file: %w", err)
	}

	contexts := make([]bindingcontext.BindingContext, 0)

	err = json.NewDecoder(contextsContent).Decode(&contexts)
	if err != nil {
		return nil, fmt.Errorf("decode binding context: %w", err)
	}

	for cidx, context := range contexts {
		for sidx, snap := range context.Snapshots {
			for ridx, res := range snap {
				// if we have empty object or filter result for some reason
				if string(res.Object) == `{}` || string(res.Object) == `"{}"` {
					contexts[cidx].Snapshots[sidx][ridx].Object = nil
				}

				if string(res.FilterResult) == `{}` || string(res.FilterResult) == `"{}"` {
					contexts[cidx].Snapshots[sidx][ridx].FilterResult = nil
				}
			}
		}
	}

	return contexts, nil
}

func (r *Request) GetDependencyContainer() pkg.DependencyContainer {
	return r.dc
}

func (r *Request) loadValuesFromFile(valuesFilePath string) (map[string]any, error) {
	valuesYaml, err := os.ReadFile(valuesFilePath)
	if err != nil && os.IsNotExist(err) {
		r.logger.Debug("no values file", slog.String("file_path", valuesFilePath), slog.String("error", err.Error()))
		return nil, nil
	}
	if err != nil {
		return nil, errors.Join(err, errors.New("load values file '"+valuesFilePath+"'"))
	}

	values, err := utils.NewValuesFromBytes(valuesYaml)
	if err != nil {
		return nil, err
	}

	return values, nil
}

func (t *Transport) NewResponse() *Response {
	return &Response{
		hookName: t.hookName,

		MetricsPath:          t.MetricsPath,
		KubernetesPath:       t.KubernetesPath,
		ValuesJSONPath:       t.ValuesJSONPath,
		ConfigValuesJSONPath: t.ConfigValuesJSONPath,

		CreateFilesByYourself: t.CreateFilesByYourself,

		logger: t.logger,
	}
}

type Response struct {
	hookName string

	MetricsPath          string
	KubernetesPath       string
	ValuesJSONPath       string
	ConfigValuesJSONPath string

	CreateFilesByYourself bool

	logger *log.Logger
}

func (r *Response) Send(res *hook.HookResult) error {
	collectors := map[string]pkg.Outputer{
		r.MetricsPath:          res.Metrics,
		r.KubernetesPath:       res.ObjectPatcherOperations,
		r.ValuesJSONPath:       res.Patches[utils.MemoryValuesPatch],
		r.ConfigValuesJSONPath: res.Patches[utils.ConfigMapPatch],
	}

	for path, collector := range collectors {
		err := r.send(path, collector)
		if err != nil {
			r.logger.Error("sending output", slog.String("path", path), slog.String("error", err.Error()))
		}
	}

	return nil
}

func (r *Response) send(path string, outputer pkg.Outputer) error {
	if r.CreateFilesByYourself {
		dir := filepath.Dir(path)

		err := os.MkdirAll(dir, 0744)
		if err != nil {
			return fmt.Errorf("mkdir all: %w", err)
		}
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	err = outputer.WriteOutput(f)
	if err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}
