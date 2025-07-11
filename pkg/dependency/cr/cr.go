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

package cr

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	oci_tools "github.com/sylabs/oci-tools/pkg/mutate"
	"github.com/tidwall/gjson"

	"github.com/deckhouse/module-sdk/pkg"
)

//go:generate minimock -i Client -o cr_mock.go

const (
	defaultTimeout = 90 * time.Second
)

type Client struct {
	registryURL string
	authConfig  authn.AuthConfig
	options     *registryOptions
}

// NewClient creates container registry client using `repo` as prefix for tags passed to methods. If insecure flag is set to true, then no cert validation is performed.
// Repo example: "cr.example.com/ns/app"
func NewClient(repo string, options ...pkg.RegistryOption) (*Client, error) {
	timeout := defaultTimeout
	// make possible to rewrite timeout in runtime
	if t := os.Getenv("REGISTRY_TIMEOUT"); t != "" {
		var err error

		timeout, err = time.ParseDuration(t)
		if err != nil {
			return nil, err
		}
	}

	opts := &registryOptions{
		timeout: timeout,
	}

	for _, opt := range options {
		opt.Apply(opts)
	}

	r := &Client{
		registryURL: repo,
		options:     opts,
	}

	if !opts.withoutAuth {
		authConfig, err := readAuthConfig(repo, opts.dockerCfg)
		if err != nil {
			return nil, err
		}

		r.authConfig = authConfig
	}

	return r, nil
}

func (r *Client) Image(ctx context.Context, tag string) (v1.Image, error) {
	imageURL := r.registryURL + ":" + tag

	var nameOpts []name.Option
	if r.options.useHTTP {
		nameOpts = append(nameOpts, name.Insecure)
	}

	ref, err := name.ParseReference(imageURL, nameOpts...) // parse options available: weak validation, etc.
	if err != nil {
		return nil, err
	}

	imageOptions := make([]remote.Option, 0)

	imageOptions = append(imageOptions, remote.WithUserAgent(r.options.userAgent))
	if !r.options.withoutAuth {
		imageOptions = append(imageOptions, remote.WithAuth(authn.FromConfig(r.authConfig)))
	}

	if r.options.ca != "" {
		imageOptions = append(imageOptions, remote.WithTransport(GetHTTPTransport(r.options.ca)))
	}

	if r.options.timeout > 0 {
		// add default timeout to prevent endless request on a huge image
		ctxWTO, cancel := context.WithTimeout(ctx, r.options.timeout)
		// seems weird - yes! but we can't call cancel here, otherwise Image outside this function would be inaccessible
		go func() {
			<-ctxWTO.Done()
			cancel()
		}()

		imageOptions = append(imageOptions, remote.WithContext(ctxWTO))
	} else {
		imageOptions = append(imageOptions, remote.WithContext(ctx))
	}

	return remote.Image(
		ref,
		imageOptions...,
	)
}

func (r *Client) ListTags(ctx context.Context) ([]string, error) {
	var nameOpts []name.Option

	if r.options.useHTTP {
		nameOpts = append(nameOpts, name.Insecure)
	}

	imageOptions := make([]remote.Option, 0)

	if !r.options.withoutAuth {
		imageOptions = append(imageOptions, remote.WithAuth(authn.FromConfig(r.authConfig)))
	}

	if r.options.ca != "" {
		imageOptions = append(imageOptions, remote.WithTransport(GetHTTPTransport(r.options.ca)))
	}

	repo, err := name.NewRepository(r.registryURL, nameOpts...)
	if err != nil {
		return nil, fmt.Errorf("parsing repo %q: %w", r.registryURL, err)
	}

	if r.options.timeout > 0 {
		// add default timeout to prevent endless request on a huge amount of tags
		ctxWTO, cancel := context.WithTimeout(ctx, r.options.timeout)
		go func() {
			<-ctxWTO.Done()
			cancel()
		}()

		imageOptions = append(imageOptions, remote.WithContext(ctxWTO))
	} else {
		imageOptions = append(imageOptions, remote.WithContext(ctx))
	}

	return remote.List(repo, imageOptions...)
}

func (r *Client) Digest(ctx context.Context, tag string) (string, error) {
	image, err := r.Image(ctx, tag)
	if err != nil {
		return "", err
	}

	d, err := image.Digest()
	if err != nil {
		return "", err
	}

	return d.String(), nil
}

func readAuthConfig(repo, dockerCfgBase64 string) (authn.AuthConfig, error) {
	r, err := parse(repo)
	if err != nil {
		return authn.AuthConfig{}, err
	}

	dockerCfg, err := base64.StdEncoding.DecodeString(dockerCfgBase64)
	if err != nil {
		// if base64 decoding failed, try to use input as it is
		dockerCfg = []byte(dockerCfgBase64)
	}

	auths := gjson.Get(string(dockerCfg), "auths").Map()
	authConfig := authn.AuthConfig{}

	// The config should have at least one .auths.* entry
	for repoName, repoAuth := range auths {
		repoNameURL, err := parse(repoName)
		if err != nil {
			return authn.AuthConfig{}, err
		}

		if repoNameURL.Host == r.Host {
			err := json.Unmarshal([]byte(repoAuth.Raw), &authConfig)
			if err != nil {
				return authn.AuthConfig{}, err
			}

			return authConfig, nil
		}
	}

	return authn.AuthConfig{}, fmt.Errorf("%q credentials not found in the dockerCfg", repo)
}

func GetHTTPTransport(ca string) http.RoundTripper {
	if ca == "" {
		return http.DefaultTransport
	}

	caPool, err := x509.SystemCertPool()
	if err != nil {
		panic(fmt.Errorf("cannot get system cert pool: %w", err))
	}

	caPool.AppendCertsFromPEM([]byte(ca))

	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   defaultTimeout,
			KeepAlive: defaultTimeout,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{RootCAs: caPool},
		TLSNextProto:          make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}
}

// parse parses url without scheme://
// if we pass url without scheme ve've got url back with two leading slashes
func parse(rawURL string) (*url.URL, error) {
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		return url.ParseRequestURI(rawURL)
	}

	return url.Parse("//" + rawURL)
}

// Extract flattens the image to a single layer and returns ReadCloser for fetching the content
func Extract(image v1.Image) (io.ReadCloser, error) {
	flattenedImage, err := oci_tools.Squash(image)
	if err != nil {
		return nil, fmt.Errorf("flattening image to a single layer: %w", err)
	}

	imageLayers, err := flattenedImage.Layers()
	if err != nil {
		return nil, fmt.Errorf("getting the image's layers: %w", err)
	}

	if len(imageLayers) != 1 {
		return nil, fmt.Errorf("unexpected number of layers: %w", err)
	}

	rc, err := imageLayers[0].Uncompressed()
	if err != nil {
		return nil, fmt.Errorf("uncompress the layer: %w", err)
	}

	return rc, nil
}
