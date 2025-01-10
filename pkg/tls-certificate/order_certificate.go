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
	"time"

	"github.com/cloudflare/cfssl/helpers"
	certificatesv1 "k8s.io/api/certificates/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	csrutil "k8s.io/client-go/util/certificate/csr"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/certificate"
)

// CertificateWaitTimeoutDefault controls default amount of time we wait for certificate
// approval in one iteration.
const CertificateWaitTimeoutDefault = 1 * time.Minute

type CertificateSecret struct {
	Name string
	Crt  []byte
	Key  []byte
}

type CertificateInfo struct {
	Certificate        string `json:"certificate,omitempty"`
	Key                string `json:"key,omitempty"`
	CertificateUpdated bool   `json:"certificate_updated,omitempty"`
}

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

func ParseSecret(secret *v1.Secret) *CertificateSecret {
	cc := &CertificateSecret{
		Name: secret.Name,
	}

	if tls, ok := secret.Data["tls.crt"]; ok {
		cc.Crt = tls
	} else if client, ok := secret.Data["client.crt"]; ok {
		cc.Crt = client
	}

	if tls, ok := secret.Data["tls.key"]; ok {
		cc.Key = tls
	} else if client, ok := secret.Data["client.key"]; ok {
		cc.Key = client
	}

	return cc
}

func IssueCertificate(ctx context.Context, input *pkg.HookInput, request OrderCertificateRequest) (*CertificateInfo, error) {
	k8, err := input.DC.GetK8sClient()
	if err != nil {
		return nil, fmt.Errorf("can't init Kubernetes client: %v", err)
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
	// _ = k8.CertificatesV1().CertificateSigningRequests().Delete(context.TODO(), request.CommonName, metav1.DeleteOptions{})
	_ = k8.Delete(ctx, &certificatesv1.CertificateSigningRequest{ObjectMeta: metav1.ObjectMeta{Name: request.CommonName}})

	csrPEM, key, err := certificate.GenerateCSR(input.Logger, request.CommonName,
		certificate.WithGroups(request.Groups...),
		certificate.WithSANs(request.SANs...))
	if err != nil {
		return nil, fmt.Errorf("error generating CSR: %v", err)
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
		return nil, fmt.Errorf("error creating CertificateSigningRequest: %v", err)
	}

	// Add CSR approved status.
	csr.Status.Conditions = append(csr.Status.Conditions,
		certificatesv1.CertificateSigningRequestCondition{
			Type:    certificatesv1.CertificateApproved,
			Status:  v1.ConditionTrue,
			Reason:  "HookApprove",
			Message: "This CSR was approved by a hook.",
		})

	// Approve CSR.
	err = k8.Status().Update(ctx, csr)
	if err != nil {
		return nil, fmt.Errorf("error approving of CertificateSigningRequest: %v", err)
	}

	ctxWTO, cancel := context.WithTimeout(context.Background(), request.WaitTimeout)
	defer cancel()

	// TODO: MUST REWRITE FUNCTION FOR CONTROLER RUNTIME CLIENT
	crtPEM, err := csrutil.WaitForCertificate(ctxWTO, nil, csr.Name, csr.UID)
	if err != nil {
		return nil, fmt.Errorf("%s CertificateSigningRequest was not signed: %v", request.CommonName, err)
	}

	// Delete CSR.
	_ = k8.Delete(ctx, &certificatesv1.CertificateSigningRequest{ObjectMeta: metav1.ObjectMeta{Name: request.CommonName}})

	info := CertificateInfo{Certificate: string(crtPEM), Key: string(key), CertificateUpdated: true}

	return &info, nil
}

// shouldGenerateNewCert checks that the certificate from the cluster matches the order
func shouldGenerateNewCert(cert []byte, request OrderCertificateRequest, durationLeft time.Duration) (bool, error) {
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
