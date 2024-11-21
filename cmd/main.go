package main

import (
	"io"

	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/deckhouse/module-sdk/internal/controller"
	"github.com/deckhouse/module-sdk/internal/transport/file"
	_ "github.com/deckhouse/module-sdk/registered-hooks"
)

// Целевая картина
// 1) У нас есть список хуков
// 2) Мы получаем запрос (подкладывает конфиг + биндинг контекст)
// 3) Формируем из конфигов и биндинг контекстов мапу хуков (на первоначальном этапе можно формировать только тот хук, который запросили)
// 4) Дергаем соответствующий хук
// 5) Результат отдаем бекенду (транспорт)

// проверить нейминг без цифр +
// вернуть массив конфигураций конфига +
// ApplyValuesPatch вынести в тесты? +
// прячу snapshots за интерфейсом
// metrics опции оставить (вернуть weith group если потерял)

func main() {
	Start()
}

func Start() {
	cfg := NewConfig()
	err := cfg.Parse()
	if err != nil {
		panic(err)
	}

	logger := log.NewLogger(log.Options{
		Level:  cfg.LogLevel.Level(),
		Output: io.Discard,
	})

	fConfig := file.Config{
		BindingContextPath: cfg.HookConfig.BindingContextPath,
		ValuesPath:         cfg.HookConfig.ValuesPath,
		ConfigValuesPath:   cfg.HookConfig.ConfigValuesPath,

		HookConfigPath: cfg.HookConfig.HookConfigPath,

		MetricsPath:          cfg.HookConfig.MetricsPath,
		KubernetesPath:       cfg.HookConfig.KubernetesPath,
		ValuesJSONPath:       cfg.HookConfig.ValuesJSONPath,
		ConfigValuesJSONPath: cfg.HookConfig.ConfigValuesJSONPath,
	}

	controller := controller.NewHookController(fConfig, logger.Named("hook-controller"))

	c := NewCMD(controller)

	c.Execute()
}
