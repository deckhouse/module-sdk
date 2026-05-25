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
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tlscertificate "github.com/deckhouse/module-sdk/common-hooks/tls-certificate"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/certificate"
	"github.com/deckhouse/module-sdk/testing/helpers"
)

const tenYears = (24 * time.Hour) * 365 * 10

// commonSANs is the canonical list of SAN entries used by most tests.
// It is rendered into the actual cert via DefaultSANs and the discovery
// values set up by setupDiscoveryValues.
var commonSANs = []string{
	"example-webhook",
	"example-webhook.d8-example-module",
	"example-webhook.d8-example-module.svc",
	"%CLUSTER_DOMAIN%://example-webhook.d8-example-module.svc",
	"%PUBLIC_DOMAIN%://example-webhook.d8-example-module.svc",
}

// expectedSANs is what we expect the hook to render the SANs into once
// the cluster domain ("cluster.local") and public domain template
// ("%s.127.0.0.1.sslip.io") are substituted.
var expectedSANs = []string{
	"example-webhook",
	"example-webhook.d8-example-module",
	"example-webhook.d8-example-module.svc",
	"example-webhook.d8-example-module.svc.cluster.local",
	"example-webhook.d8-example-module.svc.127.0.0.1.sslip.io",
}

// discoveryValues holds the bits of `global` that the TLS hook reads.
// Encoded as JSON so we can hand it directly to helpers.NewValuesFromJSON.
const discoveryValues = `{
  "global": {
    "discovery": {"clusterDomain": "cluster.local"},
    "modules":   {"publicDomainTemplate": "%s.127.0.0.1.sslip.io"}
  }
}`

// generateCertSnapshot generates a CA + cert pair using the provided
// expiries and returns it as a single Snapshot, ready for
// helpers.NewSnapshots().Add(InternalTLSSnapshotKey, ...).
func generateCertSnapshot(t *testing.T, caExpiry, certExpiry time.Duration) pkg.Snapshot {
	t.Helper()

	ca, err := certificate.GenerateCA("cert-name",
		certificate.WithKeyAlgo("ecdsa"),
		certificate.WithKeySize(256),
		certificate.WithCAExpiry(caExpiry))
	require.NoError(t, err)

	cert, err := certificate.GenerateSelfSignedCert("cert-name", ca,
		certificate.WithSANs(expectedSANs...),
		certificate.WithKeyAlgo("ecdsa"),
		certificate.WithKeySize(256),
		certificate.WithSigningDefaultExpiry(certExpiry),
		certificate.WithSigningDefaultUsage([]string{
			"signing",
			"key encipherment",
			"requestheader-client",
		}),
	)
	require.NoError(t, err)

	return helpers.SnapshotFromObject(cert)
}

// extractCertValues returns the CertValues that the hook wrote to the
// values store at the configured prefix. It fails the test if the value
// was not set.
func extractCertValues(t *testing.T, v pkg.PatchableValuesCollector, path string) tlscertificate.CertValues {
	t.Helper()
	patches := v.GetPatches()
	require.NotEmpty(t, patches, "expected the hook to write certificate values")

	want := "/" + replaceDots(path)
	for _, op := range patches {
		if op.Path != want {
			continue
		}
		var cv tlscertificate.CertValues
		require.NoError(t, json.Unmarshal(op.Value, &cv))
		return cv
	}
	t.Fatalf("no patch operation at %q (got %+v)", want, patches)
	return tlscertificate.CertValues{}
}

// replaceDots converts a dotted path into a JSON-Pointer style segmentation.
func replaceDots(p string) string {
	out := []byte(p)
	for i := range out {
		if out[i] == '.' {
			out[i] = '/'
		}
	}
	return string(out)
}

// runTLSHook invokes the hook with the given config against an input
// constructed from the provided snapshot and seeded discovery values.
// It returns the values collector so the caller can assert on patches.
func runTLSHook(t *testing.T, conf tlscertificate.GenSelfSignedTLSHookConf, snap pkg.Snapshot, extraValues string) pkg.PatchableValuesCollector {
	t.Helper()

	values := mergeValuesJSON(t, discoveryValues, extraValues)

	snaps := helpers.NewSnapshots()
	if snap != nil {
		snaps.Add(tlscertificate.InternalTLSSnapshotKey, snap)
	}

	in := helpers.NewInputBuilder(t).
		WithSnapshots(snaps).
		WithValues(values).
		Build()

	require.NoError(t, tlscertificate.GenSelfSignedTLS(conf)(context.Background(), in))
	return values
}

func mergeValuesJSON(t *testing.T, base, overlay string) pkg.PatchableValuesCollector {
	t.Helper()
	if overlay == "" {
		return helpers.NewValuesFromJSON(base)
	}
	var merged map[string]any
	require.NoError(t, json.Unmarshal([]byte(base), &merged))
	var ovr map[string]any
	require.NoError(t, json.Unmarshal([]byte(overlay), &ovr))
	deepMerge(merged, ovr)
	return helpers.NewValues(merged)
}

func deepMerge(dst, src map[string]any) {
	for k, v := range src {
		if cur, ok := dst[k].(map[string]any); ok {
			if next, ok := v.(map[string]any); ok {
				deepMerge(cur, next)
				continue
			}
		}
		dst[k] = v
	}
}

// defaultConf returns a baseline GenSelfSignedTLSHookConf that produces a
// cert at d8-example-module.internal.webhookCert. Each test starts from
// this and tweaks the fields under test.
func defaultConf() tlscertificate.GenSelfSignedTLSHookConf {
	return tlscertificate.GenSelfSignedTLSHookConf{
		CN:                   "cert-name",
		TLSSecretName:        "secret-webhook-cert",
		Namespace:            "some-namespace",
		SANs:                 tlscertificate.DefaultSANs(commonSANs),
		FullValuesPathPrefix: "d8-example-module.internal.webhookCert",
	}
}

// assertCertHasExpectedSANs checks that the resulting CertValues contain
// a non-empty CA / Crt / Key and that the cert renders the expected SANs.
func assertCertHasExpectedSANs(t *testing.T, cv tlscertificate.CertValues) {
	t.Helper()
	assert.NotEmpty(t, cv.CA)
	assert.NotEmpty(t, cv.Crt)
	assert.NotEmpty(t, cv.Key)

	cert, err := certificate.ParseCertificate([]byte(cv.Crt))
	require.NoError(t, err)
	assert.Equal(t, expectedSANs, cert.DNSNames)
}

// =============================================================================
// Tests
// =============================================================================

func TestJQFilterTLS(t *testing.T) {
	const rawSecret = `{
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

	cert := new(certificate.Certificate)
	require.NoError(t, helpers.JQRunOnString(
		context.Background(),
		tlscertificate.JQFilterTLS,
		rawSecret,
		cert,
	))

	assert.Equal(t, "some-key", string(cert.Key))
	assert.Equal(t, "some-crt", string(cert.Cert))
	assert.Equal(t, "some-ca", string(cert.CA))
}

func TestInternalTLSConfig_Valid(t *testing.T) {
	require.NoError(t, tlscertificate.GenSelfSignedTLSConfig(tlscertificate.GenSelfSignedTLSHookConf{}).Validate())
}

func TestGenSelfSignedTLS_NoCertificateInSnapshot(t *testing.T) {
	conf := defaultConf()
	conf.BeforeHookCheck = func(_ *pkg.HookInput) bool { return true }
	conf.CAExpiryDuration = 5 * 365 * 24 * time.Hour
	conf.CertExpiryDuration = 365 * 24 * time.Hour

	values := runTLSHook(t, conf, nil, "")

	cv := extractCertValues(t, values, conf.FullValuesPathPrefix)
	assertCertHasExpectedSANs(t, cv)

	ca, err := certificate.ParseCertificate([]byte(cv.CA))
	require.NoError(t, err)
	assert.True(t, ca.IsCA)
	assert.Equal(t, 5*365*24*time.Hour, ca.NotAfter.Sub(ca.NotBefore))

	cert, err := certificate.ParseCertificate([]byte(cv.Crt))
	require.NoError(t, err)
	assert.Equal(t, ca.Subject, cert.Issuer)
	assert.Equal(t, 365*24*time.Hour, cert.NotAfter.Sub(cert.NotBefore))
}

func TestGenSelfSignedTLS_ActualCertificateInSnapshot(t *testing.T) {
	snap := generateCertSnapshot(t, tenYears, tenYears)

	values := runTLSHook(t, defaultConf(), snap, "")

	cv := extractCertValues(t, values, defaultConf().FullValuesPathPrefix)
	assertCertHasExpectedSANs(t, cv)
}

func TestGenSelfSignedTLS_OutdatedCertificateInSnapshot(t *testing.T) {
	// Generate cert with 1h validity; outdated threshold is 2h, so it's "outdated".
	snap := generateCertSnapshot(t, tenYears, time.Hour)

	conf := defaultConf()
	conf.CertOutdatedDuration = 2 * time.Hour

	values := runTLSHook(t, conf, snap, "")

	cv := extractCertValues(t, values, conf.FullValuesPathPrefix)
	assertCertHasExpectedSANs(t, cv)
}

func TestGenSelfSignedTLS_OutdatedCAInSnapshot(t *testing.T) {
	snap := generateCertSnapshot(t, time.Hour, tenYears)

	conf := defaultConf()
	conf.CAOutdatedDuration = 2 * time.Hour

	values := runTLSHook(t, conf, snap, "")

	cv := extractCertValues(t, values, conf.FullValuesPathPrefix)
	assertCertHasExpectedSANs(t, cv)
}

func TestGenSelfSignedTLS_WrongCAExpiry(t *testing.T) {
	snap := generateCertSnapshot(t, tenYears, tenYears)

	conf := defaultConf()
	conf.CAExpiryDuration = 365 * 24 * time.Hour // 1 year

	values := runTLSHook(t, conf, snap, "")

	cv := extractCertValues(t, values, conf.FullValuesPathPrefix)
	assertCertHasExpectedSANs(t, cv)

	ca, err := certificate.ParseCertificate([]byte(cv.CA))
	require.NoError(t, err)
	assert.Equal(t, 365*24*time.Hour, ca.NotAfter.Sub(ca.NotBefore))
}

func TestGenSelfSignedTLS_WrongCertExpiry(t *testing.T) {
	snap := generateCertSnapshot(t, tenYears, tenYears)

	conf := defaultConf()
	conf.CertExpiryDuration = 365 * 24 * time.Hour

	values := runTLSHook(t, conf, snap, "")

	cv := extractCertValues(t, values, conf.FullValuesPathPrefix)
	assertCertHasExpectedSANs(t, cv)

	cert, err := certificate.ParseCertificate([]byte(cv.Crt))
	require.NoError(t, err)
	assert.Equal(t, 365*24*time.Hour, cert.NotAfter.Sub(cert.NotBefore))
}

func TestGenSelfSignedTLS_CommonCAWithCustomFieldName(t *testing.T) {
	ca, err := certificate.GenerateCA("cert-name",
		certificate.WithKeyAlgo("ecdsa"),
		certificate.WithKeySize(256),
		certificate.WithCAExpiry(tenYears))
	require.NoError(t, err)

	// CA is stored at custom field name "cert" instead of default "crt".
	overlay, err := json.Marshal(map[string]any{
		"global": map[string]any{
			"internal": map[string]any{
				"modules": map[string]any{
					"kubeRBACProxyCA": map[string]any{
						"cert": string(ca.Cert),
						"key":  string(ca.Key),
					},
				},
			},
		},
	})
	require.NoError(t, err)

	conf := tlscertificate.GenSelfSignedTLSHookConf{
		CN:                   "cert-name",
		TLSSecretName:        "secret-kube-rbac-proxy-cert",
		Namespace:            "some-namespace",
		SANs:                 tlscertificate.DefaultSANs([]string{"example-svc"}),
		FullValuesPathPrefix: "d8-example-module.internal.kubeRBACProxyCert",
		CommonCAValuesPath:   "global.internal.modules.kubeRBACProxyCA",
		CommonCACertField:    "cert",
	}

	values := runTLSHook(t, conf, nil, string(overlay))

	cv := extractCertValues(t, values, conf.FullValuesPathPrefix)
	assert.NotEmpty(t, cv.CA)
	assert.NotEmpty(t, cv.Crt)
	assert.NotEmpty(t, cv.Key)

	parsedCert, err := certificate.ParseCertificate([]byte(cv.Crt))
	require.NoError(t, err)
	assert.Equal(t, "cert-name", parsedCert.Issuer.CommonName)
}
