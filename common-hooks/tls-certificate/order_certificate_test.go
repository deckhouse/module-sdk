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

package tlscertificate_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tlscertificate "github.com/deckhouse/module-sdk/common-hooks/tls-certificate"
	"github.com/deckhouse/module-sdk/pkg/certificate"
	"github.com/deckhouse/module-sdk/testing/helpers"
)

// secretWithKeys is a small helper that returns a JSON-encoded TLS Secret
// whose `data.<keyPrefix>.crt`, `data.<keyPrefix>.key`, and `data.ca.crt`
// fields are pre-populated with base64 of the expected plaintext values.
func secretWithKeys(keyPrefix string) string {
	switch keyPrefix {
	case "tls":
		return `{
  "apiVersion": "v1",
  "data": {
    "ca.crt":  "c29tZS1jYQ==",
    "tls.crt": "c29tZS1jcnQ=",
    "tls.key": "c29tZS1rZXk="
  },
  "kind": "Secret",
  "metadata": {"name": "some-cert", "namespace": "some-ns"},
  "type": "kubernetes.io/tls"
}`
	case "client":
		return `{
  "apiVersion": "v1",
  "data": {
    "ca.crt":     "c29tZS1jYQ==",
    "client.crt": "c29tZS1jcnQ=",
    "client.key": "c29tZS1rZXk="
  },
  "kind": "Secret",
  "metadata": {"name": "some-cert", "namespace": "some-ns"},
  "type": "kubernetes.io/tls"
}`
	}
	panic("unsupported prefix")
}

func TestJQFilterApplyCertificateSecret(t *testing.T) {
	cases := []struct {
		name      string
		keyPrefix string
	}{
		{name: "tls keys", keyPrefix: "tls"},
		{name: "client keys", keyPrefix: "client"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cert := new(certificate.Certificate)

			require.NoError(t, helpers.JQRunOnString(
				context.Background(),
				tlscertificate.JQFilterApplyCertificateSecret,
				secretWithKeys(tc.keyPrefix),
				cert,
			))

			assert.Equal(t, "some-cert", cert.Name)
			assert.Equal(t, "some-key", string(cert.Key))
			assert.Equal(t, "some-crt", string(cert.Cert))
		})
	}
}

func TestCertificateHandlerConfig_IsValid(t *testing.T) {
	require.NoError(t, tlscertificate.CertificateHandlerConfig([]string{}, []string{}).Validate())
}
