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
	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/certificate"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	"github.com/deckhouse/module-sdk/pkg/registry"
	certificatesv1 "k8s.io/api/certificates/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CertificateWaitTimeoutDefault controls default amount of time we wait for certificate
// approval in one iteration.
const (
	CertificateWaitTimeoutDefault = 1 * time.Minute
	OrderSertificateSnapshotKey   = "certificateSecrets"
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
    "crt": (if (.data."tls.crt" != null and .data."tls.crt" != "") then .data."tls.crt" else (if (.data."client.crt" != null and .data."client.crt" != "") then .data."client.crt" else null end) end)
}`

// RegisterOrderCertificateHookEM must be used for external modules
func RegisterOrderCertificateHookEM(requests []OrderCertificateRequest) bool {
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
				Name:              OrderSertificateSnapshotKey,
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
	}, СertificateHandler(requests))
}

func СertificateHandler(requests []OrderCertificateRequest) func(ctx context.Context, input *pkg.HookInput) error {
	return func(ctx context.Context, input *pkg.HookInput) error {
		return certificateHandlerWithRequests(ctx, input, requests)
	}
}

func certificateHandlerWithRequests(ctx context.Context, input *pkg.HookInput, requests []OrderCertificateRequest) error {
	publicDomainTemplate := input.Values.Get("global.modules.publicDomainTemplate").String()
	clusterDomain := input.Values.Get("global.discovery.clusterDomain").String()

	for _, originalRequest := range requests {
		request := originalRequest.DeepCopy()

		// Convert cluster domain and public domain sans
		for index, san := range request.SANs {
			switch {
			case strings.HasPrefix(san, publicDomainPrefix) && publicDomainTemplate != "":
				request.SANs[index] = getPublicDomainSAN(publicDomainTemplate, san)

			case strings.HasPrefix(san, clusterDomainPrefix) && clusterDomain != "":
				request.SANs[index] = getClusterDomainSAN(clusterDomain, san)
			}
		}

		valueName := fmt.Sprintf("%s.%s", request.ModuleName, request.ValueName)

		certSecrets, err := objectpatch.UnmarshalToStruct[certificate.Certificate](input.Snapshots, "certificateSecrets")
		if err != nil {
			return fmt.Errorf("unmarshal to struct: %w", err)
		}

		if len(certSecrets) > 0 {
			var secret *certificate.Certificate

			for _, snapSecret := range certSecrets {
				if snapSecret.Name == request.SecretName {
					secret = &snapSecret

					break
				}
			}

			if secret != nil && len(secret.Cert) > 0 && len(secret.Key) > 0 {
				// Check that certificate is not expired and has the same order request
				genNew, err := shouldGenerateNewCert(secret.Cert, request, time.Hour*24*15)
				if err != nil {
					return fmt.Errorf("should generate new cert: %w", err)
				}

				if !genNew {
					info := CertificateInfo{Certificate: string(secret.Cert), Key: string(secret.Key)}

					input.Values.Set(valueName, info)

					continue
				}
			}
		}

		info, err := IssueCertificate(ctx, input, request)
		if err != nil {
			return fmt.Errorf("issue certificate: %w", err)
		}

		input.Values.Set(valueName, info)
	}
	return nil
}

// shouldGenerateNewCert checks that the certificate from the cluster matches the order
func shouldGenerateNewCert(cert []byte, request OrderCertificateRequest, durationLeft time.Duration) (bool, error) {
	c, err := helpers.ParseCertificatePEM(cert)
	if err != nil {
		return false, fmt.Errorf("certificate cannot parsed: %w", err)
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

type CertificateInfo struct {
	Certificate        string `json:"certificate,omitempty"`
	Key                string `json:"key,omitempty"`
	CertificateUpdated bool   `json:"certificate_updated,omitempty"`
}

func IssueCertificate(ctx context.Context, input *pkg.HookInput, request OrderCertificateRequest) (*CertificateInfo, error) {
	k8, err := input.DC.GetK8sClient()
	if err != nil {
		return nil, fmt.Errorf("can't init Kubernetes client: %w", err)
	}

	if request.WaitTimeout == 0 {
		request.WaitTimeout = CertificateWaitTimeoutDefault
	}

	if len(request.Usages) == 0 {
		request.Usages = []certificatesv1.KeyUsage{
			certificatesv1.UsageDigitalSignature,
			certificatesv1.UsageKeyEncipherment,
			certificatesv1.UsageClientAuth,
		}
	}

	if request.SignerName == "" {
		request.SignerName = certificatesv1.KubeAPIServerClientSignerName
	}

	// Delete existing CSR from the cluster.
	err = k8.Delete(ctx, &certificatesv1.CertificateSigningRequest{ObjectMeta: metav1.ObjectMeta{Name: request.CommonName}})
	if client.IgnoreNotFound(err) != nil {
		input.Logger.Warn("delete existing CSR from the cluster", log.Err(err))
	}

	csrPEM, key, err := certificate.GenerateCSR(
		request.CommonName,
		certificate.WithGroups(request.Groups...),
		certificate.WithSANs(request.SANs...))
	if err != nil {
		return nil, fmt.Errorf("error generating CSR: %w", err)
	}

	// Create new CSR in the cluster.
	csr := &certificatesv1.CertificateSigningRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CertificateSigningRequest",
			APIVersion: "certificates.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: request.CommonName,
		},
		Spec: certificatesv1.CertificateSigningRequestSpec{
			Request:           csrPEM,
			Usages:            request.Usages,
			SignerName:        request.SignerName,
			ExpirationSeconds: request.ExpirationSeconds,
		},
	}

	// Create CSR.
	err = k8.Create(ctx, csr)
	if err != nil {
		return nil, fmt.Errorf("error creating CertificateSigningRequest: %w", err)
	}

	// Add CSR approved status.
	csr.Status.Conditions = append(csr.Status.Conditions,
		certificatesv1.CertificateSigningRequestCondition{
			Type:    certificatesv1.CertificateApproved,
			Status:  v1.ConditionTrue,
			Reason:  "HookApprove",
			Message: "This CSR was approved by a hook.",
		},
	)

	// Approve CSR.
	err = k8.Status().Update(ctx, csr)
	if err != nil {
		return nil, fmt.Errorf("error approving of CertificateSigningRequest: %w", err)
	}

	ctxWTO, cancel := context.WithTimeout(context.Background(), request.WaitTimeout)
	defer cancel()

	crtPEM, err := certificate.WaitForCertificate(ctxWTO, k8, csr.Name, csr.UID)
	if err != nil {
		return nil, fmt.Errorf("%s CertificateSigningRequest was not signed: %w", request.CommonName, err)
	}

	// Delete CSR.
	err = k8.Delete(ctx, &certificatesv1.CertificateSigningRequest{ObjectMeta: metav1.ObjectMeta{Name: request.CommonName}})
	if client.IgnoreNotFound(err) != nil {
		input.Logger.Warn("delete CSR", log.Err(err))
	}

	info := CertificateInfo{
		Certificate:        string(crtPEM),
		Key:                string(key),
		CertificateUpdated: true,
	}

	return &info, nil
}
