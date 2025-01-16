/*
Copyright 2022 Flant JSC

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

package ensure_crds

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/deckhouse/module-sdk/pkg"
	crdinstaller "github.com/deckhouse/module-sdk/pkg/crd-installer"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

var defaultLabels = map[string]string{
	crdinstaller.LabelHeritage: "deckhouse",
}

func RegisterEnsureCRDsHookEM(crdsGlob string) bool {
	return registry.RegisterFunc(&pkg.HookConfig{
		OnStartup: &pkg.OrderedConfig{Order: 5},
	}, EnsureCRDsHandler(crdsGlob))
}

func EnsureCRDsHandler(crdsGlob string) func(ctx context.Context, input *pkg.HookInput) error {
	return func(ctx context.Context, input *pkg.HookInput) error {
		err := EnsureCRDs(ctx, input, crdsGlob)
		if err != nil {
			input.Logger.Error("ensure_crds failed", log.Err(err))

			return fmt.Errorf("ensure crds: %w", err)
		}

		return nil
	}
}

func EnsureCRDs(ctx context.Context, input *pkg.HookInput, crdsGlob string) error {
	client, err := input.DC.GetK8sClient()
	if err != nil {
		return fmt.Errorf("get k8s client: %w", err)
	}

	cp, err := newCRDsInstaller(client, crdsGlob)
	if err != nil {
		return fmt.Errorf("new crd installer: %w", err)
	}

	err = cp.Run(ctx)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	return nil
}

// newCRDsInstaller creates new installer for CRDs
// crdsGlob example: "/deckhouse/modules/002-deckhouse/crds/*.yaml"
func newCRDsInstaller(client pkg.KubernetesClient, crdsGlob string) (*crdinstaller.CRDsInstaller, error) {
	crds, err := filepath.Glob(crdsGlob)
	if err != nil {
		return nil, fmt.Errorf("glob %q: %w", crdsGlob, err)
	}

	return crdinstaller.NewCRDsInstaller(
		client.Dynamic(),
		crds,
		crdinstaller.WithExtraLabels(defaultLabels),
		crdinstaller.WithFileFilter(func(crdFilePath string) bool {
			return !strings.HasPrefix(filepath.Base(crdFilePath), "doc-")
		}),
	), nil
}
