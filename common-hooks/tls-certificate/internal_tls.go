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

package tlscertificate

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/cloudflare/cfssl/csr"
	certificatesv1 "k8s.io/api/certificates/v1"
	"k8s.io/utils/net"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/certificate"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

const (
	caExpiryDuration     = (24 * time.Hour) * 365 * 10 // 10 years
	caOutdatedDuration   = (24 * time.Hour) * 365 / 2  // 6 month, just enough to renew CA certificate
	certExpiryDuration   = (24 * time.Hour) * 365 * 10 // 10 years
	certOutdatedDuration = (24 * time.Hour) * 365 / 2  // 6 month, just enough to renew certificate

	InternalTLSSnapshotKey = "secret"
)

type GenSelfSignedTLSHookConf struct {
	// SANs function which returns list of domain to include into cert. Use DefaultSANs helper
	SANs SANsGenerator

	// CN - Certificate common Name
	// often it is module name
	CN string

	// Namespace - namespace for TLS secret
	Namespace string
	// TLSSecretName - TLS secret name
	// secret must be TLS secret type https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets
	// CA certificate MUST set to ca.crt key
	TLSSecretName string

	// Usages specifies valid usage contexts for keys.
	// See: https://tools.ietf.org/html/rfc5280#section-4.2.1.3
	//      https://tools.ietf.org/html/rfc5280#section-4.2.1.12
	Usages []certificatesv1.KeyUsage

	// certificate encryption algorithm
	// Can be one of: "rsa", "ecdsa", "ed25519"
	// Default: "ecdsa"
	KeyAlgorithm string

	// certificate encryption algorith key size
	// The KeySize must match the KeyAlgorithm (more info: https://github.com/cloudflare/cfssl/blob/cb0a0a3b9daf7ba477e106f2f013dd68267f0190/csr/csr.go#L108)
	// Default: 256 bit
	KeySize int

	// FullValuesPathPrefix - prefix full path to store CA certificate TLS private key and cert
	// full paths will be
	//   FullValuesPathPrefix + .ca  - CA certificate
	//   FullValuesPathPrefix + .crt - TLS private key
	//   FullValuesPathPrefix + .key - TLS certificate
	// Example: FullValuesPathPrefix =  'prometheusMetricsAdapter.internal.adapter'
	// Values to store:
	// prometheusMetricsAdapter.internal.adapter.ca
	// prometheusMetricsAdapter.internal.adapter.crt
	// prometheusMetricsAdapter.internal.adapter.key
	// Data in values store as plain text
	// In helm templates you need use `b64enc` function to encode
	FullValuesPathPrefix string

	// BeforeHookCheck runs check function before hook execution. Function should return boolean 'continue' value
	// if return value is false - hook will stop its execution
	// if return value is true - hook will continue
	BeforeHookCheck func(input *pkg.HookInput) bool

	// CommonCA - full path to store CA certificate TLS private key and cert
	// full path will be
	//   CommonCAValuesPath
	// Example: CommonCAValuesPath =  'commonCaPath'
	// Values to store:
	// commonCaPath.key
	// commonCaPath.crt
	// Data in values store as plain text
	// In helm templates you need use `b64enc` function to encode
	CommonCAValuesPath string
	// Canonical name (CN) of common CA certificate.
	// If not specified (empty), then (if no CA cert already generated) using CN property of this struct
	CommonCACanonicalName string

	// CAExpiryDuration is duration of CA certificate validity
	// if not specified (zero value), then using default value
	CAExpiryDuration time.Duration
	// CAOutdatedDuration defines how long before expiration
	// a CA certificate is considered outdated and should be renewed.
	// If not specified (zero value), then the default value is used.
	CAOutdatedDuration time.Duration
	// CertExpiryDuration is duration of certificate validity
	// if not specified (zero value), then using default value
	CertExpiryDuration time.Duration
	// CertOutdatedDuration defines how long before expiration
	// a certificate is considered outdated and should be renewed.
	// If not specified (zero value), then the default value is used.
	CertOutdatedDuration time.Duration
}

func (gss GenSelfSignedTLSHookConf) Path() string {
	return strings.TrimSuffix(gss.FullValuesPathPrefix, ".")
}

func (gss GenSelfSignedTLSHookConf) CommonCAPath() string {
	return strings.TrimSuffix(gss.CommonCAValuesPath, ".")
}

// SANsGenerator function for generating sans
type SANsGenerator func(input *pkg.HookInput) []string

var JQFilterTLS = `{
    "key": .data."tls.key",
    "crt": .data."tls.crt",
    "ca": .data."ca.crt"
}`

// RegisterInternalTLSHookEM must be used for external modules
//
// Register hook which save tls cert in values from secret.
// If secret is not created hook generate CA with long expired time
// and generate tls cert for passed domains signed with generated CA.
// That CA cert and TLS cert and private key MUST save in secret with helm.
// Otherwise, every d8 restart will generate new tls cert.
// Tls cert also has long expired time same as CA 87600h == 10 years.
// Therese tls cert often use for in cluster https communication
// with service which order tls
// Clients need to use CA cert for verify connection
func RegisterInternalTLSHookEM(conf GenSelfSignedTLSHookConf) bool {
	return registry.RegisterFunc(GenSelfSignedTLSConfig(conf), GenSelfSignedTLS(conf))
}

func GenSelfSignedTLSConfig(conf GenSelfSignedTLSHookConf) *pkg.HookConfig {
	return &pkg.HookConfig{
		OnBeforeHelm: &pkg.OrderedConfig{Order: 5},
		Kubernetes: []pkg.KubernetesConfig{
			{
				Name:       InternalTLSSnapshotKey,
				APIVersion: "v1",
				Kind:       "Secret",
				NamespaceSelector: &pkg.NamespaceSelector{
					NameSelector: &pkg.NameSelector{
						MatchNames: []string{conf.Namespace},
					},
				},
				NameSelector: &pkg.NameSelector{
					MatchNames: []string{conf.TLSSecretName},
				},
				JqFilter: JQFilterTLS,
			},
		},
		Schedule: []pkg.ScheduleConfig{
			{
				Name:    "internalTLSSchedule",
				Crontab: "42 4 * * *",
			},
		},
	}
}

type SelfSignedCertValues struct {
	CA                 *certificate.Authority
	CN                 string
	KeyAlgorithm       string
	KeySize            int
	SANs               []string
	Usages             []string
	CAExpireDuration   time.Duration
	CertExpireDuration time.Duration
}

func GenSelfSignedTLS(conf GenSelfSignedTLSHookConf) func(ctx context.Context, input *pkg.HookInput) error {
	var usages []string

	if conf.Usages == nil {
		usages = []string{
			"signing",
			"key encipherment",
			"requestheader-client",
		}
	} else {
		for _, v := range conf.Usages {
			usages = append(usages, string(v))
		}
	}

	keyAlgorithm := conf.KeyAlgorithm
	keySize := conf.KeySize

	if len(keyAlgorithm) == 0 {
		keyAlgorithm = "ecdsa"
	}

	if keySize < 128 {
		keySize = 256
	}

	// some fool-proof validation
	keyReq := csr.KeyRequest{A: keyAlgorithm, S: keySize}
	algo := keyReq.SigAlgo()
	if algo == x509.UnknownSignatureAlgorithm {
		panic(errors.New("unknown KeyAlgorithm"))
	}

	_, err := keyReq.Generate()
	if err != nil {
		panic(fmt.Errorf("bad KeySize/KeyAlgorithm combination: %w", err))
	}

	caExpiryDuration := caExpiryDuration
	if conf.CAExpiryDuration != 0 {
		caExpiryDuration = conf.CAExpiryDuration
	}
	caOutdatedDuration := caOutdatedDuration
	if conf.CAOutdatedDuration != 0 {
		caOutdatedDuration = conf.CAOutdatedDuration
	}
	certExpiryDuration := certExpiryDuration
	if conf.CertExpiryDuration != 0 {
		certExpiryDuration = conf.CertExpiryDuration
	}
	certOutdatedDuration := certOutdatedDuration
	if conf.CertOutdatedDuration != 0 {
		certOutdatedDuration = conf.CertOutdatedDuration
	}

	return func(_ context.Context, input *pkg.HookInput) error {
		if conf.BeforeHookCheck != nil {
			passed := conf.BeforeHookCheck(input)
			if !passed {
				return nil
			}
		}

		var cert *certificate.Certificate

		cn, sans := conf.CN, conf.SANs(input)

		certs, err := objectpatch.UnmarshalToStruct[certificate.Certificate](input.Snapshots, InternalTLSSnapshotKey)
		if err != nil {
			return fmt.Errorf("unmarshal to struct: %w", err)
		}

		var auth *certificate.Authority

		mustGenerate := false

		useCommonCA := conf.CommonCAValuesPath != ""

		// 1) get and validate common ca
		// 2) if not valid:
		// 2.1) regenerate common ca
		// 2.2) save new common ca in values
		// 2.3) mark certificates to regenerate
		if useCommonCA {
			auth, err = getCommonCA(input, conf.CommonCAPath(), caExpiryDuration, caOutdatedDuration)
			if err != nil {
				commonCACanonicalName := conf.CommonCACanonicalName

				if len(commonCACanonicalName) == 0 {
					commonCACanonicalName = conf.CN
				}

				auth, err = certificate.GenerateCA(
					commonCACanonicalName,
					certificate.WithKeyAlgo(keyAlgorithm),
					certificate.WithKeySize(keySize),
					certificate.WithCAExpiry(caExpiryDuration))
				if err != nil {
					return fmt.Errorf("generate ca: %w", err)
				}

				input.Values.Set(conf.CommonCAPath(), auth)

				mustGenerate = true
			}
		}

		// if no certificate - regenerate
		if len(certs) == 0 {
			mustGenerate = true
		}

		// 1) take first certificate
		// 2) check certificate ca outdated
		// 3) if using common CA - compare cert CA and common CA (if different - mark outdated)
		// 4) check certificate outdated
		// 5) if CA or cert outdated - regenerate
		if len(certs) > 0 {
			// Certificate is in the snapshot => load it.
			cert = &certs[0]

			// update certificate if less than 6 month left. We create certificate for 10 years, so it looks acceptable
			// and we don't need to create Crontab schedule
			caOutdated, err := isOutdatedCA(cert.CA, caExpiryDuration, caOutdatedDuration)
			if err != nil {
				input.Logger.Warn("is outdated ca", log.Err(err))
			}

			// if common ca and cert ca is not equal - regenerate cert
			if useCommonCA && !slices.Equal(auth.Cert, cert.CA) {
				input.Logger.Warn("common ca is not equal cert ca")

				caOutdated = true
			}

			certOutdated, err := isIrrelevantCert(cert.Cert, sans, certExpiryDuration, certOutdatedDuration)
			if err != nil {
				input.Logger.Warn("is irrelevant cert", log.Err(err))
			}

			// In case of errors, both these flags are false to avoid regeneration loop for the
			// certificate.
			mustGenerate = caOutdated || certOutdated
		}

		if mustGenerate {
			cert, err = generateNewSelfSignedTLS(SelfSignedCertValues{
				CA:                 auth,
				CN:                 cn,
				KeyAlgorithm:       keyAlgorithm,
				KeySize:            keySize,
				SANs:               sans,
				Usages:             usages,
				CAExpireDuration:   caExpiryDuration,
				CertExpireDuration: certExpiryDuration,
			})

			if err != nil {
				return fmt.Errorf("generate new self signed tls: %w", err)
			}
		}

		input.Values.Set(conf.Path(), convCertToValues(cert))

		return nil
	}
}

type CertValues struct {
	CA  string `json:"ca"`
	Crt string `json:"crt"`
	Key string `json:"key"`
}

// The certificate mapping "cert" -> "crt". We are migrating to "crt" naming for certificates
// in values.
func convCertToValues(cert *certificate.Certificate) CertValues {
	return CertValues{
		CA:  string(cert.CA),
		Crt: string(cert.Cert),
		Key: string(cert.Key),
	}
}

var ErrCertificateIsNotFound = errors.New("certificate is not found")
var ErrCAIsInvalidOrOutdated = errors.New("ca is invalid or outdated")

func getCommonCA(input *pkg.HookInput, valKey string, caExpiryDuration, caOutdatedDuration time.Duration) (*certificate.Authority, error) {
	auth := new(certificate.Authority)

	ca, ok := input.Values.GetOk(valKey)
	if !ok {
		return nil, ErrCertificateIsNotFound
	}

	err := json.Unmarshal([]byte(ca.String()), auth)
	if err != nil {
		return nil, err
	}

	outdated, err := isOutdatedCA(auth.Cert, caExpiryDuration, caOutdatedDuration)
	if err != nil {
		input.Logger.Error("is outdated ca", log.Err(err))

		return nil, err
	}

	if !outdated {
		return auth, nil
	}

	return nil, ErrCAIsInvalidOrOutdated
}

// generateNewSelfSignedTLS
//
// if you pass ca - it will be used to sign new certificate
// if pass nil ca - it will be generate to sign new certificate
func generateNewSelfSignedTLS(input SelfSignedCertValues) (*certificate.Certificate, error) {
	if input.CA == nil {
		var err error

		input.CA, err = certificate.GenerateCA(
			input.CN,
			certificate.WithKeyAlgo(input.KeyAlgorithm),
			certificate.WithKeySize(input.KeySize),
			certificate.WithCAExpiry(input.CAExpireDuration))
		if err != nil {
			return nil, fmt.Errorf("generate ca: %w", err)
		}
	}

	cert, err := certificate.GenerateSelfSignedCert(
		input.CN,
		input.CA,
		certificate.WithSANs(input.SANs...),
		certificate.WithKeyAlgo(input.KeyAlgorithm),
		certificate.WithKeySize(input.KeySize),
		certificate.WithSigningDefaultExpiry(input.CertExpireDuration),
		certificate.WithSigningDefaultUsage(input.Usages),
	)
	if err != nil {
		return nil, fmt.Errorf("generate ca: %w", err)
	}

	return cert, nil
}

// check certificate duration and SANs list
func isIrrelevantCert(certData []byte, desiredSANSs []string, certExpiryDuration, certOutdatedDuration time.Duration) (bool, error) {
	cert, err := certificate.ParseCertificate(certData)
	if err != nil {
		return false, fmt.Errorf("parse certificate: %w", err)
	}

	if cert.NotAfter.Sub(cert.NotBefore) != certExpiryDuration {
		return true, nil
	}

	if time.Until(cert.NotAfter) < certOutdatedDuration {
		return true, nil
	}

	var dnsNames, ipAddrs []string
	for _, san := range desiredSANSs {
		if net.IsIPv4String(san) {
			ipAddrs = append(ipAddrs, san)
		} else {
			dnsNames = append(dnsNames, san)
		}
	}

	if !arraysAreEqual(dnsNames, cert.DNSNames) {
		return true, nil
	}

	if len(ipAddrs) > 0 {
		certIPs := make([]string, 0, len(cert.IPAddresses))
		for _, cip := range cert.IPAddresses {
			certIPs = append(certIPs, cip.String())
		}

		if !arraysAreEqual(ipAddrs, certIPs) {
			return true, nil
		}
	}

	return false, nil
}

func isOutdatedCA(ca []byte, caExpiryDuration, caOutdatedDuration time.Duration) (bool, error) {
	// Issue a new certificate if there is no CA in the secret.
	// Without CA it is not possible to validate the certificate.
	if len(ca) == 0 {
		return true, nil
	}

	cert, err := certificate.ParseCertificate(ca)
	if err != nil {
		return false, fmt.Errorf("parse certificate: %w", err)
	}

	if cert.NotAfter.Sub(cert.NotBefore) != caExpiryDuration {
		return true, nil
	}

	if time.Until(cert.NotAfter) < caOutdatedDuration {
		return true, nil
	}

	return false, nil
}

// DefaultSANs helper to generate list of sans for certificate
// you can also use helpers:
//
//	ClusterDomainSAN(value) to generate sans with respect of cluster domain (e.g.: "app.default.svc" with "cluster.local" value will give: app.default.svc.cluster.local
//	PublicDomainSAN(value)
func DefaultSANs(sans []string) SANsGenerator {
	return func(input *pkg.HookInput) []string {
		res := make([]string, 0, len(sans))

		clusterDomain := input.Values.Get("global.discovery.clusterDomain").String()
		publicDomainTemplate := input.Values.Get("global.modules.publicDomainTemplate").String()

		for _, san := range sans {
			switch {
			case strings.HasPrefix(san, publicDomainPrefix) && publicDomainTemplate != "":
				san = getPublicDomainSAN(publicDomainTemplate, san)

			case strings.HasPrefix(san, clusterDomainPrefix) && clusterDomain != "":
				san = getClusterDomainSAN(clusterDomain, san)
			}

			res = append(res, san)
		}

		return res
	}
}
