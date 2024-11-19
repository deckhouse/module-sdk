package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/deckhouse/deckhouse/pkg/log"
	service "github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/hook"
	"github.com/deckhouse/module-sdk/pkg/kubernetes"
	"github.com/deckhouse/module-sdk/pkg/metric"
	"github.com/deckhouse/module-sdk/pkg/registry"
	"github.com/deckhouse/module-sdk/pkg/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func main() {
	cfg := NewConfig()
	err := cfg.Parse()
	if err != nil {
		panic(err)
	}

	logger := log.NewLogger(log.Options{
		Level: cfg.LogLevel.Level(),
	})

	registry.Registry().SetLogLevel(cfg.LogLevel)
	h := registry.Registry().Hooks()
	for _, hook := range h {
		fmt.Println(hook.Name)
		fmt.Println(hook.Path)

		logger = logger.Named("hook-controller")

		res, err := hook.Execute()
		if err != nil {
			logger.Error("execute", slog.String("error", err.Error()))
		}

		collectors := map[string]service.Outputer{
			cfg.HookConfig.MetricsPath:          res.Metrics,
			cfg.HookConfig.KubernetesPath:       res.ObjectPatcherOperations,
			cfg.HookConfig.ConfigValuesJSONPath: res.Patches[utils.ConfigMapPatch],
			cfg.HookConfig.ValuesJSONPath:       res.Patches[utils.MemoryValuesPatch],
		}

		for path, collector := range collectors {
			func() {
				dir := filepath.Dir(path)

				err := os.MkdirAll(dir, 0744)
				if err != nil {
					logger.Error("mkdir all", slog.String("error", err.Error()))

					return
				}

				f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
				defer func() {
					_ = f.Close()
				}()
				if err != nil {
					logger.Error("mkdir all", slog.String("error", err.Error()))

					return
				}

				err = collector.WriteOutput(f)
				if err != nil {
					logger.Error("write output", slog.String("error", err.Error()))

					return
				}
			}()
		}
	}
}

func NewHookInput(logger *log.Logger) *hook.HookInput {
	mc := metric.NewCollector()
	mc.Add("kek", 2.0, map[string]string{"first": "first-label"})
	mc.Add("kek", 2.0, map[string]string{"second": "second-label"})

	pc := kubernetes.NewObjectPatchCollector()
	pc.Create(&unstructured.Unstructured{Object: map[string]interface{}{"version": "v01", "metadata": "meta"}})

	pv, _ := hook.NewPatchableValues(nil)
	pvc, _ := hook.NewPatchableValues(nil)

	return &hook.HookInput{
		ConfigValues:    pvc,
		Values:          pv,
		MetricCollector: mc,
		PatchCollector:  pc,
		Logger:          logger,
	}
}
