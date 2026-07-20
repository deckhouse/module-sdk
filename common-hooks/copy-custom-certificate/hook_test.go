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

package copycustomcertificate_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	copycustomcertificate "github.com/deckhouse/module-sdk/common-hooks/copy-custom-certificate"
	tlscertificate "github.com/deckhouse/module-sdk/common-hooks/tls-certificate"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/certificate"
	"github.com/deckhouse/module-sdk/testing/framework"
	"github.com/deckhouse/module-sdk/testing/helpers"
)

// tlsSecret is the canonical TLS-typed kubernetes Secret payload used by
// the JQ filters. The values are base64-encoded so the filter can decode
// them back into "some-key", "some-crt", etc.
const tlsSecret = `{
  "apiVersion": "v1",
  "data": {
    "ca.crt":  "c29tZS1jYQ==",
    "tls.crt": "c29tZS1jcnQ=",
    "tls.key": "c29tZS1rZXk="
  },
  "kind": "Secret",
  "metadata": {
    "name":      "some-cert",
    "namespace": "some-ns"
  },
  "type": "kubernetes.io/tls"
}`

const clientSecret = `{
  "apiVersion": "v1",
  "data": {
    "ca.crt":     "c29tZS1jYQ==",
    "client.crt": "c29tZS1jcnQ=",
    "client.key": "c29tZS1rZXk="
  },
  "kind": "Secret",
  "metadata": {
    "name":      "some-cert",
    "namespace": "some-ns"
  },
  "type": "kubernetes.io/tls"
}`

func TestJQFilterCustomCertificate_ParsesTLSSecret(t *testing.T) {
	cert := new(certificate.Certificate)

	require.NoError(t, helpers.JQRunOnString(
		context.Background(),
		copycustomcertificate.JQFilterCustomCertificate,
		tlsSecret,
		cert,
	))

	assert.Equal(t, "some-cert", cert.Name)
	assert.Equal(t, "some-key", string(cert.Key))
	assert.Equal(t, "some-crt", string(cert.Cert))
	assert.Equal(t, "some-ca", string(cert.CA))
}

func TestJQFilterApplyCertificateSecret_ParsesClientCertificate(t *testing.T) {
	auth := new(certificate.Certificate)

	require.NoError(t, helpers.JQRunOnString(
		context.Background(),
		tlscertificate.JQFilterApplyCertificateSecret,
		clientSecret,
		auth,
	))

	assert.Equal(t, "some-cert", auth.Name)
	assert.Equal(t, "some-key", string(auth.Key))
	assert.Equal(t, "some-crt", string(auth.Cert))
}

func TestHandler_SecretNotFound_ReturnsWarningNotError(t *testing.T) {
	cfg := &pkg.HookConfig{
		OnBeforeHelm: &pkg.OrderedConfig{Order: 10},
		Kubernetes: []pkg.KubernetesConfig{
			{
				Name:       "custom_certificates",
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
				JqFilter: copycustomcertificate.JQFilterCustomCertificate,
			},
		},
	}

	handler := copycustomcertificate.CopyCustomCertificatesHandler("testmodule")
	hec := framework.HookExecutionConfigInit(t, cfg, handler, `{}`, `{}`)

	hec.KubeStateSet(`
---
apiVersion: v1
kind: Secret
metadata:
  name: other-cert
  namespace: d8-system
  labels:
    owner: deckhouse
data:
  tls.crt: Q1JUQ1JUQ1JU
  tls.key: S0VZS0VZS0VZ
type: kubernetes.io/tls
`)

	hec.ValuesSet("testmodule.https.mode", "CustomCertificate")
	hec.ValuesSet("testmodule.https.customCertificate.secretName", "missing-cert")

	hec.RunHook()

	require.NoError(t, hec.HookError())

	assert.Equal(t, "<none>", hec.ValuesGet("testmodule.internal.customCertificateData.ca\\.crt").String())
	assert.Equal(t, "<none>", hec.ValuesGet("testmodule.internal.customCertificateData.tls\\.key").String())
	assert.Equal(t, "<none>", hec.ValuesGet("testmodule.internal.customCertificateData.tls\\.crt").String())
}

func TestHandler_SecretExists_SetsValues(t *testing.T) {
	cfg := &pkg.HookConfig{
		OnBeforeHelm: &pkg.OrderedConfig{Order: 10},
		Kubernetes: []pkg.KubernetesConfig{
			{
				Name:       "custom_certificates",
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
				JqFilter: copycustomcertificate.JQFilterCustomCertificate,
			},
		},
	}

	handler := copycustomcertificate.CopyCustomCertificatesHandler("testmodule")
	hec := framework.HookExecutionConfigInit(t, cfg, handler, `{}`, `{}`)

	hec.KubeStateSet(`
---
apiVersion: v1
kind: Secret
metadata:
  name: my-cert
  namespace: d8-system
  labels:
    owner: deckhouse
data:
  tls.crt: Q1JUQ1JUQ1JU
  tls.key: S0VZS0VZS0VZ
  ca.crt: Q0FDQUNBQ0E=
type: kubernetes.io/tls
`)

	hec.ValuesSet("testmodule.https.mode", "CustomCertificate")
	hec.ValuesSet("testmodule.https.customCertificate.secretName", "my-cert")

	hec.RunHook()

	require.NoError(t, hec.HookError())

	assert.Equal(t, "CACACACA", hec.ValuesGet("testmodule.internal.customCertificateData.ca\\.crt").String())
	assert.Equal(t, "KEYKEYKEY", hec.ValuesGet("testmodule.internal.customCertificateData.tls\\.key").String())
	assert.Equal(t, "CRTCRTCRT", hec.ValuesGet("testmodule.internal.customCertificateData.tls\\.crt").String())
}

func TestHandler_NonCustomCertificateMode_RemovesValues(t *testing.T) {
	cfg := &pkg.HookConfig{
		OnBeforeHelm: &pkg.OrderedConfig{Order: 10},
		Kubernetes: []pkg.KubernetesConfig{
			{
				Name:       "custom_certificates",
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
				JqFilter: copycustomcertificate.JQFilterCustomCertificate,
			},
		},
	}

	handler := copycustomcertificate.CopyCustomCertificatesHandler("testmodule")
	hec := framework.HookExecutionConfigInit(t, cfg, handler, `{}`, `{}`)

	hec.KubeStateSet(``)
	hec.ValuesSet("testmodule.https.mode", "CertManager")

	hec.RunHook()

	require.NoError(t, hec.HookError())
	assert.False(t, hec.ValuesGet("testmodule.internal.customCertificateData").Exists())
}
