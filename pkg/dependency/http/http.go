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

package http

//go:generate minimock -i Client -o http_mock.go

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/deckhouse/module-sdk/pkg"
)

func NewClient(options ...pkg.HTTPOption) *http.Client {
	opts := &httpOptions{
		timeout: 10 * time.Second,
	}

	for _, opt := range options {
		opt.Apply(opts)
	}

	dialer := &net.Dialer{
		Timeout: opts.timeout,
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: opts.insecure,
	}

	if !opts.insecure {
		caPool, err := x509.SystemCertPool()
		if err != nil {
			panic(fmt.Errorf("cannot get system cert pool: %w", err))
		}

		for _, ca := range opts.additionalTLSCA {
			caPool.AppendCertsFromPEM(ca)
		}

		tlsConf.RootCAs = caPool
	}

	if opts.tlsServerName != "" {
		tlsConf.ServerName = opts.tlsServerName
	}

	tr := &http.Transport{
		TLSClientConfig:       tlsConf,
		IdleConnTimeout:       5 * time.Minute,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext:           dialer.DialContext,
		Dial:                  dialer.Dial,
	}

	return &http.Client{
		Timeout:   opts.timeout,
		Transport: tr,
	}
}

func SetBearerToken(req *http.Request, token string) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
}

func GetKubeToken() (string, error) {
	if os.Getenv("D8_IS_TESTS_ENVIRONMENT") == "true" {
		return "kube-auth-test-token", nil
	}

	content, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func SetKubeAuthToken(req *http.Request) error {
	token, err := GetKubeToken()
	if err != nil {
		return err
	}

	SetBearerToken(req, token)

	return nil
}
