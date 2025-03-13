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

package certificate

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/cloudflare/cfssl/cli/genkey"
	"github.com/cloudflare/cfssl/config"
	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/helpers"
	"github.com/cloudflare/cfssl/signer"
	"github.com/cloudflare/cfssl/signer/local"
)

type Certificate struct {
	Name string `json:"name,omitempty"`

	Key  []byte `json:"key"`
	Cert []byte `json:"crt"`
	CA   []byte `json:"ca"`
}

type SigningOption func(signing *config.Signing)

func WithSigningDefaultExpiry(expiry time.Duration) SigningOption {
	return func(signing *config.Signing) {
		signing.Default.Expiry = expiry
		signing.Default.ExpiryString = expiry.String()
	}
}

func WithSigningDefaultUsage(usage []string) SigningOption {
	return func(signing *config.Signing) {
		signing.Default.Usage = usage
	}
}

func GenerateSelfSignedCert(cn string, ca *Authority, options ...any) (*Certificate, error) {
	if ca == nil {
		return nil, errors.New("ca is nil")
	}

	request := &csr.CertificateRequest{
		CN: cn,
		KeyRequest: &csr.KeyRequest{
			A: "ecdsa",
			S: 256,
		},
	}

	for _, option := range options {
		if f, ok := option.(Option); ok {
			f(request)
		}
	}

	// Catch cfssl output and show it only if error is occurred.
	var buf bytes.Buffer
	logWriter := log.Writer()

	log.SetOutput(&buf)
	defer log.SetOutput(logWriter)

	g := &csr.Generator{Validator: genkey.Validator}
	csrBytes, key, err := g.ProcessRequest(request)
	if err != nil {
		return nil, fmt.Errorf("process request: %w", err)
	}

	req := signer.SignRequest{Request: string(csrBytes)}

	parsedCa, err := helpers.ParseCertificatePEM(ca.Cert)
	if err != nil {
		return nil, fmt.Errorf("parse certificate pem: %w", err)
	}

	priv, err := helpers.ParsePrivateKeyPEM(ca.Key)
	if err != nil {
		return nil, fmt.Errorf("parse private key pem: %w", err)
	}

	signingConfig := &config.Signing{
		Default: config.DefaultConfig(),
	}

	for _, option := range options {
		if f, ok := option.(SigningOption); ok {
			f(signingConfig)
		}
	}

	s, err := local.NewSigner(priv, parsedCa, signer.DefaultSigAlgo(priv), signingConfig)
	if err != nil {
		return nil, fmt.Errorf("new signer: %w", err)
	}

	cert, err := s.Sign(req)
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}

	return &Certificate{
		CA:   ca.Cert,
		Key:  key,
		Cert: cert,
	}, nil
}
