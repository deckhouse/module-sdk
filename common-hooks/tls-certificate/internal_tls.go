package tlscertificate

import (
	"context"
	"fmt"
	"time"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/certificate"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	"github.com/deckhouse/module-sdk/pkg/registry"
	tlscertificate "github.com/deckhouse/module-sdk/pkg/tls-certificate"
	"k8s.io/utils/net"
)

const (
	caExpiryDurationStr  = "87600h"                    // 10 years
	certExpiryDuration   = (24 * time.Hour) * 365 * 10 // 10 years
	certOutdatedDuration = (24 * time.Hour) * 365 / 2  // 6 month, just enough to renew certificate

	// certificate encryption algorithm
	keyAlgorithm = "ecdsa"
	keySize      = 256
)

var JQFilterTLS = `{
    "key": .data."tls.key",
    "cert": .data."tls.crt",
    "ca": .data."ca.crt"
}`

// RegisterInternalTLSHook
// Register hook which save tls cert in values from secret.
// If secret is not created hook generate CA with long expired time
// and generate tls cert for passed domains signed with generated CA.
// That CA cert and TLS cert and private key MUST save in secret with helm.
// Otherwise, every d8 restart will generate new tls cert.
// Tls cert also has long expired time same as CA 87600h == 10 years.
// Therese tls cert often use for in cluster https communication
// with service which order tls
// Clients need to use CA cert for verify connection
func RegisterInternalTLSHook(conf tlscertificate.GenSelfSignedTLSHookConf) bool {
	return registry.RegisterFunc(&pkg.HookConfig{
		OnBeforeHelm: &pkg.OrderedConfig{Order: 5},
		Kubernetes: []pkg.KubernetesConfig{
			{
				Name:       tlscertificate.SnapshotKey,
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
	}, genSelfSignedTLS(conf))
}

func genSelfSignedTLS(conf tlscertificate.GenSelfSignedTLSHookConf) func(ctx context.Context, input *pkg.HookInput) error {
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

	return func(ctx context.Context, input *pkg.HookInput) error {
		if conf.BeforeHookCheck != nil {
			passed := conf.BeforeHookCheck(input)
			if !passed {
				return nil
			}
		}

		var cert certificate.Certificate
		var err error

		cn, sans := conf.CN, conf.SANs(input)

		certs, err := objectpatch.UnmarshalToStruct[certificate.Certificate](input.Snapshots, tlscertificate.SnapshotKey)
		if err != nil {
			return fmt.Errorf("unmarshal to struct: %w", err)
		}

		if len(certs) == 0 {
			// No certificate in snapshot => generate a new one.
			// Secret will be updated by Helm.
			cert, err = generateNewSelfSignedTLS(input, cn, sans, usages)
			if err != nil {
				return err
			}
		} else {
			// Certificate is in the snapshot => load it.
			cert = certs[0]

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
				cert, err = generateNewSelfSignedTLS(input, cn, sans, usages)
				if err != nil {
					return err
				}
			}
		}

		input.Values.Set(conf.Path(), convCertToValues(cert))

		return nil
	}
}

type certValues struct {
	CA  string `json:"ca"`
	Crt string `json:"crt"`
	Key string `json:"key"`
}

// The certificate mapping "cert" -> "crt". We are migrating to "crt" naming for certificates
// in values.
func convCertToValues(cert certificate.Certificate) certValues {
	return certValues{
		CA:  cert.CA,
		Crt: cert.Cert,
		Key: cert.Key,
	}
}

func generateNewSelfSignedTLS(input *pkg.HookInput, cn string, sans, usages []string) (certificate.Certificate, error) {
	ca, err := certificate.GenerateCA(input.Logger,
		cn,
		certificate.WithKeyAlgo(keyAlgorithm),
		certificate.WithKeySize(keySize),
		certificate.WithCAExpiry(caExpiryDurationStr))
	if err != nil {
		return certificate.Certificate{}, err
	}

	return certificate.GenerateSelfSignedCert(input.Logger,
		cn,
		ca,
		certificate.WithSANs(sans...),
		certificate.WithKeyAlgo(keyAlgorithm),
		certificate.WithKeySize(keySize),
		certificate.WithSigningDefaultExpiry(certExpiryDuration),
		certificate.WithSigningDefaultUsage(usages),
	)
}

// check certificate duration and SANs list
func isIrrelevantCert(certData string, desiredSANSs []string) (bool, error) {
	cert, err := certificate.ParseCertificate(certData)
	if err != nil {
		return false, err
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

func isOutdatedCA(ca string) (bool, error) {
	// Issue a new certificate if there is no CA in the secret.
	// Without CA it is not possible to validate the certificate.
	if len(ca) == 0 {
		return true, nil
	}

	cert, err := certificate.ParseCertificate(ca)
	if err != nil {
		return false, err
	}

	if time.Until(cert.NotAfter) < certOutdatedDuration {
		return true, nil
	}

	return false, nil
}
