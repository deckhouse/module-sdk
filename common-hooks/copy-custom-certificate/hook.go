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

package copycustomcertificate

import (
	"context"
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/certificate"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	"github.com/deckhouse/module-sdk/pkg/registry"
	valuespatch "github.com/deckhouse/module-sdk/pkg/utils/values-patch"
)

const (
	snapshotKey = "custom_certificates"
)

var JQFilterCustomCertificate = `{
    "name": .metadata.name,
    "key": .data."tls.key",
    "crt": .data."tls.crt",
    "ca": .data."ca.crt"
}`

func RegisterHook(moduleName string) bool {
	return registry.RegisterFunc(&pkg.HookConfig{
		OnBeforeHelm: &pkg.OrderedConfig{Order: 10},
		Kubernetes: []pkg.KubernetesConfig{
			{
				Name:       snapshotKey,
				APIVersion: "v1",
				Kind:       "Secret",
				NamespaceSelector: &pkg.NamespaceSelector{
					NameSelector: &pkg.NameSelector{
						MatchNames: []string{"d8-system"},
					},
				},
				LabelSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "owner",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"helm"},
						},
					},
				},
				JqFilter: JQFilterCustomCertificate,
			},
		},
	}, copyCustomCertificatesHandler(moduleName))
}

func copyCustomCertificatesHandler(moduleName string) func(ctx context.Context, input *pkg.HookInput) error {
	return func(ctx context.Context, input *pkg.HookInput) error {
		certs, err := objectpatch.UnmarshalToStruct[certificate.Certificate](input.Snapshots, snapshotKey)
		if err != nil {
			return fmt.Errorf("unmarshal to struct: %w", err)
		}

		if len(certs) == 0 {
			input.Logger.Info("No custom certificates received, skipping setting values")

			return nil
		}

		customCertificates := make(map[string]certificate.Certificate, len(certs))
		for _, cert := range certs {
			customCertificates[cert.Name] = cert
		}

		httpsMode := valuespatch.GetHTTPSMode(input, moduleName)

		valuesPath := moduleName + ".internal.customCertificateData"
		// fmt.Sprintf("%s.https.customCertificate.secretName", moduleName)
		configValuesPath := moduleName + ".https.customCertificate.secretName"

		if httpsMode != "CustomCertificate" {
			input.Values.Remove(valuesPath)

			return nil
		}

		rawsecretName, ok := valuespatch.GetValuesFirstDefined(input, configValuesPath, "global.modules.https.customCertificate.secretName")

		secretName := rawsecretName.String()

		if !ok || secretName == "" {
			return nil
		}

		cert, ok := customCertificates[secretName]
		if !ok {
			return errors.New("custom certificate secret name is configured, but secret with this name '" + secretName + "' doesn't exist")
		}

		input.Values.Set(valuesPath, certValues{
			CA:      string(cert.CA),
			TLSKey:  string(cert.Key),
			TLSCert: string(cert.Cert),
		})

		return nil
	}
}

type certValues struct {
	CA      string `json:"ca.crt,omitempty"`
	TLSKey  string `json:"tls.key,omitempty"`
	TLSCert string `json:"tls.crt,omitempty"`
}
