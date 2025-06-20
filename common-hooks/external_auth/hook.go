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

//nolint:staticcheck
package external_auth

import (
	"context"
	"strings"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
	"github.com/deckhouse/module-sdk/pkg/utils/set"
)

type Settings struct {
	// Values path to read and store auth values
	ExternalAuthPath string
	// Where to store a flag to enabled DexAuthenticator
	DexAuthenticatorEnabledPath string
	// Options to set if Dex is enabled
	DexExternalAuth ExternalAuth
}

type ExternalAuth struct {
	AuthURL         string `json:"authURL"`
	AuthSignInURL   string `json:"authSignInURL"`
	UseBearerTokens *bool  `json:"useBearerTokens,omitempty"`
}

func (e *ExternalAuth) AuthURLWithClusterDomain(input *pkg.HookInput) string {
	clusterDomain := input.Values.Get("global.discovery.clusterDomain").String()
	return strings.ReplaceAll(e.AuthURL, "%CLUSTER_DOMAIN%", clusterDomain)
}

func RegisterHook(settings Settings) bool {
	return registry.RegisterFunc(&pkg.HookConfig{
		OnBeforeHelm: &pkg.OrderedConfig{Order: 9},
	}, wrapSetExternalAuthValues(settings))
}

func setExternalAuthValues(input *pkg.HookInput, settings Settings) error {
	configAuth, isExternalAuthInConfig := input.ConfigValues.GetOk(settings.ExternalAuthPath)

	if !set.NewFromValues(input.Values, "global.enabledModules").Has("user-authn") {
		if !isExternalAuthInConfig {
			input.Values.Remove(settings.ExternalAuthPath)
		} else {
			input.Values.Set(settings.ExternalAuthPath, configAuth.Value())
		}

		input.Values.Remove(settings.DexAuthenticatorEnabledPath)
		return nil
	}

	if !isExternalAuthInConfig {
		input.Values.Set(settings.ExternalAuthPath, ExternalAuth{
			AuthURL:         settings.DexExternalAuth.AuthURLWithClusterDomain(input),
			AuthSignInURL:   settings.DexExternalAuth.AuthSignInURL,
			UseBearerTokens: settings.DexExternalAuth.UseBearerTokens,
		})
		input.Values.Set(settings.DexAuthenticatorEnabledPath, true)
	} else {
		input.Values.Set(settings.ExternalAuthPath, configAuth.Value())
		input.Values.Remove(settings.DexAuthenticatorEnabledPath)
	}

	return nil
}

func wrapSetExternalAuthValues(settings Settings) func(ctx context.Context, input *pkg.HookInput) error {
	return func(_ context.Context, input *pkg.HookInput) error {
		return setExternalAuthValues(input, settings)
	}
}
