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

	"github.com/deckhouse/deckhouse/pkg/log"
	tlscertificate "github.com/deckhouse/module-sdk/common-hooks/tls-certificate"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/certificate"
	"github.com/deckhouse/module-sdk/pkg/jq"
	mock "github.com/deckhouse/module-sdk/testing/mock"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func Test_JQFilterTLS(t *testing.T) {
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

		q, err := jq.NewQuery(tlscertificate.JQFilterTLS)
		assert.NoError(t, err)

		res, err := q.FilterStringObject(context.Background(), rawSecret)
		assert.NoError(t, err)

		cert := new(certificate.Certificate)
		err = json.NewDecoder(bytes.NewBufferString(res.String())).Decode(cert)
		assert.NoError(t, err)

		assert.Equal(t, "some-key", string(cert.Key))
		assert.Equal(t, "some-crt", string(cert.Cert))
		assert.Equal(t, "some-ca", string(cert.CA))
	})
}

func Test_GenSelfSignedTLS(t *testing.T) {
	t.Run("refoncile func", func(t *testing.T) {
		dc := mock.NewDependencyContainerMock(t)

		snapshots := mock.NewSnapshotsMock(t)
		snapshots.GetMock.When(tlscertificate.InternalTLSSnapshotKey).Then(
			[]pkg.Snapshot{
				mock.NewSnapshotMock(t).UnmarhalToMock.Set(func(v any) (err error) {
					cert := v.(*certificate.Certificate)
					*cert = certificate.Certificate{}

					return nil
				}),
			},
		)

		values := mock.NewPatchableValuesCollectorMock(t)

		values.GetMock.When("global.discovery.clusterDomain").Then(gjson.Result{Str: "d8-example-module"})
		values.GetMock.When("global.modules.publicDomainTemplate").Then(gjson.Result{Str: "%.d8-example-module"})

		values.SetMock.Set(func(path string, v any) {
			assert.Equal(t, "d8-example-module.internal.webhookCert", path)

			values, ok := v.(tlscertificate.CertValues)
			assert.True(t, ok)

			assert.NotEmpty(t, values.CA)
			assert.NotEmpty(t, values.Crt)
			assert.NotEmpty(t, values.Key)
		})

		var input = &pkg.HookInput{
			Snapshots: snapshots,
			Values:    values,
			DC:        dc,
			Logger:    log.NewNop(),
		}

		config := tlscertificate.GenSelfSignedTLSHookConf{
			CN:            "secr-name",
			TLSSecretName: "secret-webhook-cert",
			Namespace:     "some-namespace",
			SANs: tlscertificate.DefaultSANs([]string{
				"example-webhook",
				"example-webhook.d8-example-module",
				"example-webhook.d8-example-module.svc",
				"example-webhook.d8-example-module.svc.cluster.local",
			}),
			FullValuesPathPrefix: "d8-example-module.internal.webhookCert",
		}

		err := tlscertificate.GenSelfSignedTLS(config)(context.Background(), input)
		assert.NoError(t, err)
	})
}
