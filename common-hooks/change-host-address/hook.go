/*
Copyright 2021 Flant JSC

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

package changehostaddress

import (
	"context"
	"fmt"

	"github.com/deckhouse/module-sdk/pkg"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	"github.com/deckhouse/module-sdk/pkg/registry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	initialHostAddressAnnotation = "node.deckhouse.io/initial-host-ip"
	snapshotKey                  = "pod"
)

type Address struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	InitialHost string `json:"initialHost"`
}

var JQFilterGetAddress = `{
    "name": .metadata.name,
    "host": .status.hostIP,
    "initialHost": (if (.metadata.annotations."node.deckhouse.io/initial-host-ip" != null and .metadata.annotations."node.deckhouse.io/initial-host-ip" != "") then .metadata.annotations."node.deckhouse.io/initial-host-ip" end)
}`

func RegisterHookEM(appName, namespace string) bool {
	return registry.RegisterFunc(&pkg.HookConfig{
		Kubernetes: []pkg.KubernetesConfig{
			{
				Name:       snapshotKey,
				APIVersion: "v1",
				Kind:       "Pod",
				NamespaceSelector: &pkg.NamespaceSelector{
					NameSelector: &pkg.NameSelector{
						MatchNames: []string{namespace},
					},
				},
				LabelSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "app",
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{appName},
						},
					},
				},
				JqFilter: JQFilterGetAddress,
			},
		},
	}, wrapChangeAddressHandler(namespace))
}

func wrapChangeAddressHandler(namespace string) func(ctx context.Context, input *pkg.HookInput) error {
	return func(ctx context.Context, input *pkg.HookInput) error {
		return changeHostAddressHandler(ctx, input, namespace)
	}
}

func changeHostAddressHandler(_ context.Context, input *pkg.HookInput, namespace string) error {
	adresses, err := objectpatch.UnmarshalToStruct[Address](input.Snapshots, snapshotKey)
	if err != nil {
		return fmt.Errorf("unmarshal to struct: %w", err)
	}

	if len(adresses) == 0 {
		return nil
	}

	for _, podAddress := range adresses {
		if podAddress.Host == "" {
			// Pod doesn't exist, we can skip it
			continue
		}

		if podAddress.InitialHost == "" {
			patch := map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						initialHostAddressAnnotation: podAddress.Host,
					},
				},
			}

			input.PatchCollector.MergePatch(patch, "v1", "Pod", namespace, podAddress.Name)

			continue
		}

		if podAddress.InitialHost != podAddress.Host {
			input.PatchCollector.Delete("v1", "Pod", namespace, podAddress.Name)
		}
	}

	return nil
}
