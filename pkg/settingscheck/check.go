/*
Copyright 2025 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package settingscheck

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/deckhouse/module-sdk/pkg"
	patchablevalues "github.com/deckhouse/module-sdk/pkg/patchable-values"
	"github.com/deckhouse/module-sdk/pkg/utils"
)

const (
	EnvSettingsPath = "SETTINGS_PATH"
)

type Result struct {
	Allow   bool   `json:"allow" yaml:"allow"`
	Warning string `json:"warning" yaml:"warning"`
}

type Warning struct {
	Message string `json:"message" yaml:"message"`
}

func (w *Warning) Error() string {
	return w.Message
}

type Input struct {
	Settings pkg.ReadOnlyValuesCollector
	DC       pkg.DependencyContainer
	Logger   pkg.Logger
}

type Check func(ctx context.Context, input Input) error

func Wrap(ctx context.Context, check Check, dc pkg.DependencyContainer, logger pkg.Logger) Result {
	if check == nil {
		return Result{
			Allow: true,
		}
	}

	path := os.Getenv(EnvSettingsPath)
	if path == "" {
		return Result{
			Allow:   false,
			Warning: fmt.Sprintf("env '%s' not set", EnvSettingsPath),
		}
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return Result{
			Allow:   false,
			Warning: fmt.Sprintf("failed to read settings: %v", err),
		}
	}

	values, err := utils.NewValuesFromBytes(raw)
	if err != nil {
		return Result{
			Allow:   false,
			Warning: fmt.Sprintf("failed to parse settings: %v", err),
		}
	}

	settings, err := patchablevalues.NewPatchableValues(values)
	if err != nil {
		return Result{
			Allow:   false,
			Warning: fmt.Sprintf("failed to parse settings: %v", err),
		}
	}

	input := Input{
		Settings: settings,
		DC:       dc,
		Logger:   logger,
	}

	err = check(ctx, input)
	if err == nil {
		return Result{
			Allow: true,
		}
	}

	warning := new(Warning)
	if errors.As(err, &warning) {
		return Result{
			Allow:   true,
			Warning: warning.Message,
		}
	}

	return Result{
		Allow:   false,
		Warning: err.Error(),
	}
}
