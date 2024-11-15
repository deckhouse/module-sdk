/*
Copyright 2024 Flant JSC

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

package extenders

import (
	"slices"

	"github.com/deckhouse/module-sdk/pkg/dependency/extenders/bootstrapped"
	"github.com/deckhouse/module-sdk/pkg/dependency/extenders/deckhouseversion"
	"github.com/deckhouse/module-sdk/pkg/dependency/extenders/kubernetesversion"
	"github.com/deckhouse/module-sdk/pkg/scheduler/extenders"
)

func IsExtendersField(field string) bool {
	return slices.Contains([]string{
		kubernetesversion.RequirementsField,
		deckhouseversion.RequirementsField,
		bootstrapped.RequirementsField,
	}, field)
}

func Extenders() []extenders.Extender {
	return []extenders.Extender{
		kubernetesversion.Instance(),
		deckhouseversion.Instance(),
		bootstrapped.Instance(),
	}
}

func AddConstraints(module string, requirements map[string]string) error {
	if len(requirements[deckhouseversion.RequirementsField]) > 0 {
		if err := deckhouseversion.Instance().AddConstraint(module, requirements[deckhouseversion.RequirementsField]); err != nil {
			return err
		}
	}
	if len(requirements[kubernetesversion.RequirementsField]) > 0 {
		if err := kubernetesversion.Instance().AddConstraint(module, requirements[kubernetesversion.RequirementsField]); err != nil {
			return err
		}
	}
	if len(requirements[bootstrapped.RequirementsField]) > 0 {
		if err := bootstrapped.Instance().AddConstraint(module, requirements[bootstrapped.RequirementsField]); err != nil {
			return err
		}
	}
	return nil
}

func DeleteConstraints(module string) {
	deckhouseversion.Instance().DeleteConstraint(module)
	kubernetesversion.Instance().DeleteConstraint(module)
	bootstrapped.Instance().DeleteConstraint(module)
}

func CheckModuleReleaseRequirements(moduleRelease string, requirements map[string]string) error {
	if len(requirements[kubernetesversion.RequirementsField]) > 0 {
		if err := kubernetesversion.Instance().ValidateRelease(moduleRelease, requirements[kubernetesversion.RequirementsField]); err != nil {
			return err
		}
	}
	if len(requirements[deckhouseversion.RequirementsField]) > 0 {
		if err := deckhouseversion.Instance().ValidateRelease(moduleRelease, requirements[deckhouseversion.RequirementsField]); err != nil {
			return err
		}
	}
	return nil
}
