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
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

// ParseCertificatesFromBase64 parsing base64 input string and return ca cert and/or verified tls.Certificate
func ParseCertificatesFromBase64(ca, crt, key string) (*x509.Certificate, *tls.Certificate, error) {
	caCert, err := generateCACert(ca)
	if err != nil {
		return nil, nil, fmt.Errorf("generate ca cert: %w", err)
	}

	clientCert, err := generateTLSCert(crt, key)
	if err != nil {
		return nil, nil, fmt.Errorf("generate tls cert: %w", err)
	}

	return caCert, clientCert, nil
}

func generateCACert(caBase64 string) (*x509.Certificate, error) {
	if caBase64 == "" {
		return nil, nil
	}

	caData, err := base64.StdEncoding.DecodeString(caBase64)
	if err != nil {
		return nil, fmt.Errorf("base64 decode string ca: %w", err)
	}

	block, _ := pem.Decode(caData)
	if block == nil {
		return nil, ErrBlockNotFound
	}

	if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
		return nil, fmt.Errorf("not valid ca certificate")
	}

	return x509.ParseCertificate(block.Bytes)
}

func generateTLSCert(crt, key string) (*tls.Certificate, error) {
	if crt == "" || key == "" {
		return nil, nil
	}

	certData, err := base64.StdEncoding.DecodeString(crt)
	if err != nil {
		return nil, fmt.Errorf("base64 decode string crt: %w", err)
	}
	keyData, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("base64 decode string key: %w", err)
	}

	cert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		return nil, fmt.Errorf("tls x509 key pair: %w", err)
	}
	return &cert, nil
}

// ParseCertificatesFromPEM parsing PEM input strings and return ca cert and/or verified tls.Certificate
func ParseCertificatesFromPEM(ca, crt, key []byte) (*x509.Certificate, *tls.Certificate, error) {
	caCert, err := ParseCertificate(ca)
	if err != nil {
		return nil, nil, fmt.Errorf("parse certificate: %w", err)
	}

	clientCert, err := tls.X509KeyPair(crt, key)
	if err != nil {
		return nil, nil, fmt.Errorf("tls x509 key pair: %w", err)
	}

	return caCert, &clientCert, nil
}

var ErrNotValidCACertificate = errors.New("not valid ca certificate")
var ErrBlockNotFound = errors.New("block not found")

// ParseCertificate parse x509 certificate PEM encoded
func ParseCertificate(crt []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(crt)
	if block == nil {
		return nil, ErrBlockNotFound
	}

	if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
		return nil, ErrNotValidCACertificate
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("x509 parse certificate: %w", err)
	}

	return cert, nil
}
