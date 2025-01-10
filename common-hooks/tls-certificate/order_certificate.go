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

package tlscertificate

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cloudflare/cfssl/helpers"
	certificatesv1 "k8s.io/api/certificates/v1"

	"github.com/deckhouse/module-sdk/pkg"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	"github.com/deckhouse/module-sdk/pkg/registry"
	tlscertificate "github.com/deckhouse/module-sdk/pkg/tls-certificate"
)

type OrderCertificateRequest struct {
	Namespace  string
	SecretName string
	CommonName string
	SANs       []string
	Groups     []string
	Usages     []certificatesv1.KeyUsage
	SignerName string

	ValueName   string
	ModuleName  string
	WaitTimeout time.Duration

	ExpirationSeconds *int32
}

func (r *OrderCertificateRequest) DeepCopy() OrderCertificateRequest {
	newR := OrderCertificateRequest{
		Namespace:  r.Namespace,
		SecretName: r.SecretName,
		CommonName: r.CommonName,
		SignerName: r.SignerName,
		SANs:       append(make([]string, 0, len(r.SANs)), r.SANs...),
		Groups:     append(make([]string, 0, len(r.Groups)), r.Groups...),
		Usages:     append(make([]certificatesv1.KeyUsage, 0, len(r.Usages)), r.Usages...),

		ValueName:   r.ValueName,
		ModuleName:  r.ModuleName,
		WaitTimeout: r.WaitTimeout,
	}
	return newR
}

var JQFilterApplyCertificateSecret = `{
    "name": .metadata.name,
    "key": (if (.data."tls.key" != null and .data."tls.key" != "") then .data."tls.key" else (if (.data."client.key" != null and .data."client.key" != "") then .data."client.key" else null end) end),
    "cert": (if (.data."tls.crt" != null and .data."tls.crt" != "") then .data."tls.crt" else (if (.data."client.crt" != null and .data."client.crt" != "") then .data."client.crt" else null end) end)
}`

func RegisterOrderCertificateHook(requests []tlscertificate.OrderCertificateRequest) bool {
	var namespaces []string
	var secretNames []string
	for _, request := range requests {
		namespaces = append(namespaces, request.Namespace)
		secretNames = append(secretNames, request.SecretName)
	}
	return registry.RegisterFunc(&pkg.HookConfig{
		OnBeforeHelm: &pkg.OrderedConfig{
			Order: 5,
		},
		Kubernetes: []pkg.KubernetesConfig{
			{
				Name:              "certificateSecrets",
				APIVersion:        "v1",
				Kind:              "Secret",
				NamespaceSelector: &pkg.NamespaceSelector{NameSelector: &pkg.NameSelector{MatchNames: namespaces}},
				NameSelector:      &pkg.NameSelector{MatchNames: secretNames},
				JqFilter:          JQFilterApplyCertificateSecret,
			},
		},
		Schedule: []pkg.ScheduleConfig{
			{
				Name:    "certificateCheck",
				Crontab: "42 4 * * *",
			},
		},
	}, certificateHandler(requests))
}

func certificateHandler(requests []tlscertificate.OrderCertificateRequest) func(ctx context.Context, input *pkg.HookInput) error {
	return func(ctx context.Context, input *pkg.HookInput) error {
		return certificateHandlerWithRequests(ctx, input, requests)
	}
}

func certificateHandlerWithRequests(ctx context.Context, input *pkg.HookInput, requests []tlscertificate.OrderCertificateRequest) error {
	publicDomain := input.Values.Get("global.modules.publicDomainTemplate").String()
	clusterDomain := input.Values.Get("global.discovery.clusterDomain").String()

	for _, originalRequest := range requests {
		request := originalRequest.DeepCopy()

		// Convert cluster domain and public domain sans
		for index, san := range request.SANs {
			switch {
			case strings.HasPrefix(san, tlscertificate.PublicDomainPrefix) && publicDomain != "":
				request.SANs[index] = tlscertificate.GetPublicDomainSAN(san, publicDomain)

			case strings.HasPrefix(san, tlscertificate.ClusterDomainPrefix) && clusterDomain != "":
				request.SANs[index] = tlscertificate.GetClusterDomainSAN(san, clusterDomain)
			}
		}

		valueName := fmt.Sprintf("%s.%s", request.ModuleName, request.ValueName)

		certSecrets, err := objectpatch.UnmarshalToStruct[tlscertificate.CertificateSecret](input.Snapshots, "certificateSecrets")
		if err != nil {
			return fmt.Errorf("unmarshal to struct: %w", err)
		}

		if len(certSecrets) > 0 {
			var secret *tlscertificate.CertificateSecret

			for _, snapSecret := range certSecrets {
				if snapSecret.Name == request.SecretName {
					secret = &snapSecret
					break
				}
			}

			if secret != nil && len(secret.Crt) > 0 && len(secret.Key) > 0 {
				// Check that certificate is not expired and has the same order request
				genNew, err := shouldGenerateNewCert(secret.Crt, request, time.Hour*24*15)
				if err != nil {
					return err
				}
				if !genNew {
					info := tlscertificate.CertificateInfo{Certificate: string(secret.Crt), Key: string(secret.Key)}
					input.Values.Set(valueName, info)
					continue
				}
			}
		}

		info, err := tlscertificate.IssueCertificate(ctx, input, request)
		if err != nil {
			return err
		}
		input.Values.Set(valueName, info)
	}
	return nil
}

// shouldGenerateNewCert checks that the certificate from the cluster matches the order
func shouldGenerateNewCert(cert []byte, request tlscertificate.OrderCertificateRequest, durationLeft time.Duration) (bool, error) {
	c, err := helpers.ParseCertificatePEM(cert)
	if err != nil {
		return false, fmt.Errorf("certificate cannot parsed: %v", err)
	}

	if c.Subject.CommonName != request.CommonName {
		return true, nil
	}

	if !arraysAreEqual(c.Subject.Organization, request.Groups) {
		return true, nil
	}

	if !arraysAreEqual(c.DNSNames, request.SANs) {
		return true, nil
	}

	// TODO: compare usages
	// if !arraysAreEqual(c.ExtKeyUsage, request.Usages) {
	//	  return true, nil
	// }

	if time.Until(c.NotAfter) < durationLeft {
		return true, nil
	}
	return false, nil
}

func arraysAreEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sort.Strings(a)
	sort.Strings(b)

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
