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
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	tlscertificate "github.com/deckhouse/module-sdk/common-hooks/tls-certificate"
	"github.com/deckhouse/module-sdk/pkg/certificate"
	"github.com/deckhouse/module-sdk/pkg/jq"
)

func Test_JQFilterApplyCertificateSecret(t *testing.T) {
	t.Run("apply tls", func(t *testing.T) {
		const rawSecret = `
		{
	  "apiVersion": "v1",
	  "data": {
		"ca.crt": "c29tZS1jYQ==",
		"tls.crt": "c29tZS1jcnQ=",
		"tls.key": "c29tZS1rZXk="
	  },
	  "kind": "Secret",
	  "metadata": {
		"name": "some-cert",
		"namespace": "some-ns"
	  },
	  "type": "kubernetes.io/tls"
	}`

		q, err := jq.NewQuery(tlscertificate.JQFilterApplyCertificateSecret)
		assert.NoError(t, err)

		res, err := q.FilterStringObject(context.Background(), rawSecret)
		assert.NoError(t, err)

		auth := new(certificate.Certificate)
		err = json.NewDecoder(bytes.NewBufferString(res.String())).Decode(auth)
		assert.NoError(t, err)

		assert.Equal(t, "some-key", string(auth.Key))
		assert.Equal(t, "some-crt", string(auth.Cert))
		assert.Equal(t, "some-cert", auth.Name)
	})

	t.Run("apply tls from client", func(t *testing.T) {
		const rawSecret = `
		{
	  "apiVersion": "v1",
	  "data": {
		"ca.crt": "c29tZS1jYQ==",
		"client.crt": "c29tZS1jcnQ=",
		"client.key": "c29tZS1rZXk="
	  },
	  "kind": "Secret",
	  "metadata": {
		"name": "some-cert",
		"namespace": "some-ns"
	  },
	  "type": "kubernetes.io/tls"
	}`

		q, err := jq.NewQuery(tlscertificate.JQFilterApplyCertificateSecret)
		assert.NoError(t, err)

		res, err := q.FilterStringObject(context.Background(), rawSecret)
		assert.NoError(t, err)

		cert := new(certificate.Certificate)
		err = json.NewDecoder(bytes.NewBufferString(res.String())).Decode(cert)
		assert.NoError(t, err)

		assert.Equal(t, "some-key", string(cert.Key))
		assert.Equal(t, "some-crt", string(cert.Cert))
		assert.Equal(t, "some-cert", cert.Name)
	})
}

func Test_CertificateHandlerConfig(t *testing.T) {
	t.Run("config is valid", func(t *testing.T) {
		assert.NoError(t, tlscertificate.CertificateHandlerConfig([]string{}, []string{}).Validate())
	})
}
