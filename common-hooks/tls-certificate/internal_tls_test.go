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

package tlscertificate_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

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

func Test_InternalTLSConfig(t *testing.T) {
	t.Run("config is valid", func(t *testing.T) {
		assert.NoError(t, tlscertificate.GenSelfSignedTLSConfig(tlscertificate.GenSelfSignedTLSHookConf{}).Validate())
	})
}

func Test_GenSelfSignedTLS(t *testing.T) {
	t.Run("no certificate in snapshot", func(t *testing.T) {
		dc := mock.NewDependencyContainerMock(t)

		snapshots := mock.NewSnapshotsMock(t)
		snapshots.GetMock.When(tlscertificate.InternalTLSSnapshotKey).Then(
			[]pkg.Snapshot{},
		)

		values := mock.NewPatchableValuesCollectorMock(t)

		values.GetMock.When("global.discovery.clusterDomain").Then(gjson.Result{Type: gjson.String, Str: "cluster.local"})
		values.GetMock.When("global.modules.publicDomainTemplate").Then(gjson.Result{Type: gjson.String, Str: "%s.127.0.0.1.sslip.io"})

		values.SetMock.Set(func(path string, v any) {
			assert.Equal(t, "d8-example-module.internal.webhookCert", path)

			values, ok := v.(tlscertificate.CertValues)
			assert.True(t, ok)

			assert.NotEmpty(t, values.CA)
			assert.NotEmpty(t, values.Crt)
			assert.NotEmpty(t, values.Key)

			cert, err := certificate.ParseCertificate([]byte(values.Crt))
			assert.NoError(t, err)

			assert.Equal(t, []string{
				"example-webhook",
				"example-webhook.d8-example-module",
				"example-webhook.d8-example-module.svc",
				"example-webhook.d8-example-module.svc.cluster.local",
				"example-webhook.d8-example-module.svc.127.0.0.1.sslip.io",
			}, cert.DNSNames)
		})

		var input = &pkg.HookInput{
			Snapshots: snapshots,
			Values:    values,
			DC:        dc,
			Logger:    log.NewNop(),
		}

		config := tlscertificate.GenSelfSignedTLSHookConf{
			CN:            "cert-name",
			TLSSecretName: "secret-webhook-cert",
			Namespace:     "some-namespace",
			BeforeHookCheck: func(_ *pkg.HookInput) bool {
				return true
			},
			SANs: tlscertificate.DefaultSANs([]string{
				"example-webhook",
				"example-webhook.d8-example-module",
				"example-webhook.d8-example-module.svc",
				"%CLUSTER_DOMAIN%://example-webhook.d8-example-module.svc",
				"%PUBLIC_DOMAIN%://example-webhook.d8-example-module.svc",
			}),
			FullValuesPathPrefix: "d8-example-module.internal.webhookCert",
		}

		err := tlscertificate.GenSelfSignedTLS(config)(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("actual certificate in snapshot", func(t *testing.T) {
		dc := mock.NewDependencyContainerMock(t)

		snapshots := mock.NewSnapshotsMock(t)
		snapshots.GetMock.When(tlscertificate.InternalTLSSnapshotKey).Then(
			[]pkg.Snapshot{
				mock.NewSnapshotMock(t).UnmarhalToMock.Set(func(v any) (err error) {
					ca, err := certificate.GenerateCA(
						"cert-name",
						certificate.WithKeyAlgo("ecdsa"),
						certificate.WithKeySize(256),
						certificate.WithCAExpiry("87600h"))

					assert.NoError(t, err)

					cert, err := certificate.GenerateSelfSignedCert(
						"cert-name",
						ca,
						certificate.WithSANs([]string{
							"example-webhook",
							"example-webhook.d8-example-module",
							"example-webhook.d8-example-module.svc",
							"example-webhook.d8-example-module.svc.cluster.local",
							"example-webhook.d8-example-module.svc.127.0.0.1.sslip.io",
						}...),
						certificate.WithKeyAlgo("ecdsa"),
						certificate.WithKeySize(256),
						certificate.WithSigningDefaultExpiry((24*time.Hour)*365*10),
						certificate.WithSigningDefaultUsage([]string{
							"signing",
							"key encipherment",
							"requestheader-client",
						}),
					)

					assert.NoError(t, err)

					value := v.(*certificate.Certificate)
					*value = *cert

					return nil
				}),
			},
		)

		values := mock.NewPatchableValuesCollectorMock(t)

		values.GetMock.When("global.discovery.clusterDomain").Then(gjson.Result{Type: gjson.String, Str: "cluster.local"})
		values.GetMock.When("global.modules.publicDomainTemplate").Then(gjson.Result{Type: gjson.String, Str: "%s.127.0.0.1.sslip.io"})

		values.SetMock.Set(func(path string, v any) {
			assert.Equal(t, "d8-example-module.internal.webhookCert", path)

			values, ok := v.(tlscertificate.CertValues)
			assert.True(t, ok)

			assert.NotEmpty(t, values.CA)
			assert.NotEmpty(t, values.Crt)
			assert.NotEmpty(t, values.Key)

			cert, err := certificate.ParseCertificate([]byte(values.Crt))
			assert.NoError(t, err)

			assert.Equal(t, []string{
				"example-webhook",
				"example-webhook.d8-example-module",
				"example-webhook.d8-example-module.svc",
				"example-webhook.d8-example-module.svc.cluster.local",
				"example-webhook.d8-example-module.svc.127.0.0.1.sslip.io",
			}, cert.DNSNames)
		})

		var input = &pkg.HookInput{
			Snapshots: snapshots,
			Values:    values,
			DC:        dc,
			Logger:    log.NewNop(),
		}

		config := tlscertificate.GenSelfSignedTLSHookConf{
			CN:            "cert-name",
			TLSSecretName: "secret-webhook-cert",
			Namespace:     "some-namespace",
			SANs: tlscertificate.DefaultSANs([]string{
				"example-webhook",
				"example-webhook.d8-example-module",
				"example-webhook.d8-example-module.svc",
				"%CLUSTER_DOMAIN%://example-webhook.d8-example-module.svc",
				"%PUBLIC_DOMAIN%://example-webhook.d8-example-module.svc",
			}),
			FullValuesPathPrefix: "d8-example-module.internal.webhookCert",
		}

		err := tlscertificate.GenSelfSignedTLS(config)(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("outdated certificate in snapshot", func(t *testing.T) {
		dc := mock.NewDependencyContainerMock(t)

		snapshots := mock.NewSnapshotsMock(t)
		snapshots.GetMock.When(tlscertificate.InternalTLSSnapshotKey).Then(
			[]pkg.Snapshot{
				mock.NewSnapshotMock(t).UnmarhalToMock.Set(func(v any) (err error) {
					ca, err := certificate.GenerateCA(
						"cert-name",
						certificate.WithKeyAlgo("ecdsa"),
						certificate.WithKeySize(256),
						certificate.WithCAExpiry("87600h"))

					assert.NoError(t, err)

					cert, err := certificate.GenerateSelfSignedCert(
						"cert-name",
						ca,
						certificate.WithSANs([]string{
							"example-webhook",
							"example-webhook.d8-example-module",
							"example-webhook.d8-example-module.svc",
							"example-webhook.d8-example-module.svc.cluster.local",
							"example-webhook.d8-example-module.svc.127.0.0.1.sslip.io",
						}...),
						certificate.WithKeyAlgo("ecdsa"),
						certificate.WithKeySize(256),
						certificate.WithSigningDefaultExpiry(1*time.Hour),
						certificate.WithSigningDefaultUsage([]string{
							"signing",
							"key encipherment",
							"requestheader-client",
						}),
					)

					assert.NoError(t, err)

					value := v.(*certificate.Certificate)
					*value = *cert

					return nil
				}),
			},
		)

		values := mock.NewPatchableValuesCollectorMock(t)

		values.GetMock.When("global.discovery.clusterDomain").Then(gjson.Result{Type: gjson.String, Str: "cluster.local"})
		values.GetMock.When("global.modules.publicDomainTemplate").Then(gjson.Result{Type: gjson.String, Str: "%s.127.0.0.1.sslip.io"})

		values.SetMock.Set(func(path string, v any) {
			assert.Equal(t, "d8-example-module.internal.webhookCert", path)

			values, ok := v.(tlscertificate.CertValues)
			assert.True(t, ok)

			assert.NotEmpty(t, values.CA)
			assert.NotEmpty(t, values.Crt)
			assert.NotEmpty(t, values.Key)

			cert, err := certificate.ParseCertificate([]byte(values.Crt))
			assert.NoError(t, err)

			assert.Equal(t, []string{
				"example-webhook",
				"example-webhook.d8-example-module",
				"example-webhook.d8-example-module.svc",
				"example-webhook.d8-example-module.svc.cluster.local",
				"example-webhook.d8-example-module.svc.127.0.0.1.sslip.io",
			}, cert.DNSNames)
		})

		var input = &pkg.HookInput{
			Snapshots: snapshots,
			Values:    values,
			DC:        dc,
			Logger:    log.NewNop(),
		}

		config := tlscertificate.GenSelfSignedTLSHookConf{
			CN:            "cert-name",
			TLSSecretName: "secret-webhook-cert",
			Namespace:     "some-namespace",
			SANs: tlscertificate.DefaultSANs([]string{
				"example-webhook",
				"example-webhook.d8-example-module",
				"example-webhook.d8-example-module.svc",
				"%CLUSTER_DOMAIN%://example-webhook.d8-example-module.svc",
				"%PUBLIC_DOMAIN%://example-webhook.d8-example-module.svc",
			}),
			FullValuesPathPrefix: "d8-example-module.internal.webhookCert",
		}

		err := tlscertificate.GenSelfSignedTLS(config)(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("outdated ca in snapshot", func(t *testing.T) {
		dc := mock.NewDependencyContainerMock(t)

		snapshots := mock.NewSnapshotsMock(t)
		snapshots.GetMock.When(tlscertificate.InternalTLSSnapshotKey).Then(
			[]pkg.Snapshot{
				mock.NewSnapshotMock(t).UnmarhalToMock.Set(func(v any) (err error) {
					ca, err := certificate.GenerateCA(
						"cert-name",
						certificate.WithKeyAlgo("ecdsa"),
						certificate.WithKeySize(256),
						certificate.WithCAExpiry("1h"))

					assert.NoError(t, err)

					cert, err := certificate.GenerateSelfSignedCert(
						"cert-name",
						ca,
						certificate.WithSANs([]string{
							"example-webhook",
							"example-webhook.d8-example-module",
							"example-webhook.d8-example-module.svc",
							"example-webhook.d8-example-module.svc.cluster.local",
							"example-webhook.d8-example-module.svc.127.0.0.1.sslip.io",
						}...),
						certificate.WithKeyAlgo("ecdsa"),
						certificate.WithKeySize(256),
						certificate.WithSigningDefaultExpiry((24*time.Hour)*365*10),
						certificate.WithSigningDefaultUsage([]string{
							"signing",
							"key encipherment",
							"requestheader-client",
						}),
					)

					assert.NoError(t, err)

					value := v.(*certificate.Certificate)
					*value = *cert

					return nil
				}),
			},
		)

		values := mock.NewPatchableValuesCollectorMock(t)

		values.GetMock.When("global.discovery.clusterDomain").Then(gjson.Result{Type: gjson.String, Str: "cluster.local"})
		values.GetMock.When("global.modules.publicDomainTemplate").Then(gjson.Result{Type: gjson.String, Str: "%s.127.0.0.1.sslip.io"})

		values.SetMock.Set(func(path string, v any) {
			assert.Equal(t, "d8-example-module.internal.webhookCert", path)

			values, ok := v.(tlscertificate.CertValues)
			assert.True(t, ok)

			assert.NotEmpty(t, values.CA)
			assert.NotEmpty(t, values.Crt)
			assert.NotEmpty(t, values.Key)

			cert, err := certificate.ParseCertificate([]byte(values.Crt))
			assert.NoError(t, err)

			assert.Equal(t, []string{
				"example-webhook",
				"example-webhook.d8-example-module",
				"example-webhook.d8-example-module.svc",
				"example-webhook.d8-example-module.svc.cluster.local",
				"example-webhook.d8-example-module.svc.127.0.0.1.sslip.io",
			}, cert.DNSNames)
		})

		var input = &pkg.HookInput{
			Snapshots: snapshots,
			Values:    values,
			DC:        dc,
			Logger:    log.NewNop(),
		}

		config := tlscertificate.GenSelfSignedTLSHookConf{
			CN:            "cert-name",
			TLSSecretName: "secret-webhook-cert",
			Namespace:     "some-namespace",
			SANs: tlscertificate.DefaultSANs([]string{
				"example-webhook",
				"example-webhook.d8-example-module",
				"example-webhook.d8-example-module.svc",
				"%CLUSTER_DOMAIN%://example-webhook.d8-example-module.svc",
				"%PUBLIC_DOMAIN%://example-webhook.d8-example-module.svc",
			}),
			FullValuesPathPrefix: "d8-example-module.internal.webhookCert",
		}

		err := tlscertificate.GenSelfSignedTLS(config)(context.Background(), input)
		assert.NoError(t, err)
	})
}
