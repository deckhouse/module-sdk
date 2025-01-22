package tlscertificate

import (
	"context"
	"fmt"
	"strings"
	"time"

	certificatesv1 "k8s.io/api/certificates/v1"
	"k8s.io/utils/net"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/certificate"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

const (
	caExpiryDurationStr  = "87600h"                    // 10 years
	certExpiryDuration   = (24 * time.Hour) * 365 * 10 // 10 years
	certOutdatedDuration = (24 * time.Hour) * 365 / 2  // 6 month, just enough to renew certificate

	// certificate encryption algorithm
	keyAlgorithm           = "ecdsa"
	keySize                = 256
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
}

func (gss GenSelfSignedTLSHookConf) Path() string {
	return strings.TrimSuffix(gss.FullValuesPathPrefix, ".")
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

		if len(certs) == 0 {
			// No certificate in snapshot => generate a new one.
			// Secret will be updated by Helm.
			cert, err = generateNewSelfSignedTLS(cn, sans, usages)
			if err != nil {
				return fmt.Errorf("generate new self signed tls: %w", err)
			}
		} else {
			// Certificate is in the snapshot => load it.
			cert = &certs[0]

			// update certificate if less than 6 month left. We create certificate for 10 years, so it looks acceptable
			// and we don't need to create Crontab schedule
			caOutdated, err := isOutdatedCA(cert.CA)
			if err != nil {
				input.Logger.Error(err.Error())
			}

			certOutdated, err := isIrrelevantCert(cert.Cert, sans)
			if err != nil {
				input.Logger.Error(err.Error())
			}

			// In case of errors, both these flags are false to avoid regeneration loop for the
			// certificate.
			if caOutdated || certOutdated {
				cert, err = generateNewSelfSignedTLS(cn, sans, usages)
				if err != nil {
					return fmt.Errorf("re generate new self signed tls: %w", err)
				}
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

func generateNewSelfSignedTLS(cn string, sans, usages []string) (*certificate.Certificate, error) {
	ca, err := certificate.GenerateCA(
		cn,
		certificate.WithKeyAlgo(keyAlgorithm),
		certificate.WithKeySize(keySize),
		certificate.WithCAExpiry(caExpiryDurationStr))
	if err != nil {
		return nil, fmt.Errorf("generate ca: %w", err)
	}

	cert, err := certificate.GenerateSelfSignedCert(
		cn,
		*ca,
		certificate.WithSANs(sans...),
		certificate.WithKeyAlgo(keyAlgorithm),
		certificate.WithKeySize(keySize),
		certificate.WithSigningDefaultExpiry(certExpiryDuration),
		certificate.WithSigningDefaultUsage(usages),
	)
	if err != nil {
		return nil, fmt.Errorf("generate ca: %w", err)
	}

	return cert, nil
}

// check certificate duration and SANs list
func isIrrelevantCert(certData []byte, desiredSANSs []string) (bool, error) {
	cert, err := certificate.ParseCertificate(certData)
	if err != nil {
		return false, fmt.Errorf("parse certificate: %w", err)
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

func isOutdatedCA(ca []byte) (bool, error) {
	// Issue a new certificate if there is no CA in the secret.
	// Without CA it is not possible to validate the certificate.
	if len(ca) == 0 {
		return true, nil
	}

	cert, err := certificate.ParseCertificate(ca)
	if err != nil {
		return false, fmt.Errorf("parse certificate: %w", err)
	}

	if time.Until(cert.NotAfter) < certOutdatedDuration {
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
