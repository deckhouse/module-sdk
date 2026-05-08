package hookinfolder_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/framework"

	subfolder "example-module/subfolder"
)

// openapiDir is the path to this module's OpenAPI schemas, resolved
// relative to this test file. The schemas live next to `hooks/` in
// `examples/example-module/openapi/`.
const openapiDir = "../../openapi"

// TestOpenAPI_DefaultsSeedValuesStore demonstrates the WithOpenAPIDir
// option: the framework reads the module's OpenAPI schemas and pre-seeds
// the values / config-values stores with every `default:` declared by
// them, exactly like addon-operator does in production.
//
// This is the most realistic starting point for a hook test in a module
// that ships an `openapi/` directory: hooks running under the framework
// observe the same defaults a real cluster would see.
func TestOpenAPI_DefaultsSeedValuesStore(t *testing.T) {
	cfg := &pkg.HookConfig{
		Metadata:     pkg.HookMetadata{Name: "openapi-defaults-demo"},
		OnBeforeHelm: &pkg.OrderedConfig{Order: 1},
	}

	// A small hook that branches on the `https.mode` config value. In a
	// real test you would point this at one of your handlers; here we
	// keep it inline so the assertion is right next to the schema fields
	// it exercises.
	handler := func(_ context.Context, input *pkg.HookInput) error {
		mode := input.Values.Get("https.mode").String()
		input.Values.Set("module.runtime.https.enabled", mode != "Disabled")
		input.Values.Set("module.runtime.replicas", input.ConfigValues.Get("replicas").Int())
		return nil
	}

	hec := framework.NewHookExecutionConfig(t, cfg, handler,
		framework.WithOpenAPIDir(openapiDir),
	)
	hec.RunHook()
	require.NoError(t, hec.HookError())

	// Defaults from openapi/config-values.yaml.
	assert.EqualValues(t, 1, hec.ConfigValuesGet("replicas").Int(),
		"replicas default should come from config-values.yaml")
	assert.Equal(t, "Disabled", hec.ConfigValuesGet("https.mode").String())
	assert.Equal(t, "letsencrypt",
		hec.ConfigValuesGet("https.certManager.clusterIssuerName").String())

	// Values inherit config-values via x-extend, plus get their own
	// schema-only fields.
	assert.Equal(t, "Disabled", hec.ValuesGet("https.mode").String(),
		"values must inherit https.mode default via x-extend")
	assert.True(t, hec.ValuesGet("internal.golangVersions").Exists(),
		"values.yaml-only defaults (internal.*) must be present")

	// And the hook's own writes land on top of the schema defaults.
	assert.False(t, hec.ValuesGet("module.runtime.https.enabled").Bool())
	assert.EqualValues(t, 1, hec.ValuesGet("module.runtime.replicas").Int())
}

// TestOpenAPI_UserOverridesWin pins the override semantics: anything
// passed via WithInitialValues / WithInitialConfigValues is deep-merged
// on top of the schema defaults, so the test author always wins.
func TestOpenAPI_UserOverridesWin(t *testing.T) {
	cfg := &pkg.HookConfig{
		Metadata:     pkg.HookMetadata{Name: "openapi-user-overrides"},
		OnBeforeHelm: &pkg.OrderedConfig{Order: 1},
	}

	handler := func(_ context.Context, input *pkg.HookInput) error {
		input.Values.Set("module.runtime.https.enabled",
			input.Values.Get("https.mode").String() != "Disabled")
		return nil
	}

	hec := framework.NewHookExecutionConfig(t, cfg, handler,
		framework.WithOpenAPIDir(openapiDir),
		framework.WithInitialConfigValues(`
replicas: 5
https:
  mode: CertManager
  certManager:
    clusterIssuerName: my-issuer
`),
		framework.WithInitialValues(`
https:
  mode: CertManager
`),
	)
	hec.RunHook()
	require.NoError(t, hec.HookError())

	// Explicit overrides win.
	assert.EqualValues(t, 5, hec.ConfigValuesGet("replicas").Int())
	assert.Equal(t, "CertManager", hec.ConfigValuesGet("https.mode").String())
	assert.Equal(t, "my-issuer",
		hec.ConfigValuesGet("https.certManager.clusterIssuerName").String())

	// Untouched defaults survive: customCertificate.secretName is set by
	// the schema and was not overridden.
	assert.Equal(t, "false",
		hec.ConfigValuesGet("https.customCertificate.secretName").String())

	// And the hook's branch on https.mode picked up the override.
	assert.True(t, hec.ValuesGet("module.runtime.https.enabled").Bool())
}

// TestOpenAPI_WithExistingHook demonstrates that WithOpenAPIDir composes
// with the actual hooks shipped by this module. The values hook writes
// to `some.path.to.field.*` paths that are independent of the schema, so
// schema defaults and hook output coexist.
func TestOpenAPI_WithExistingHook(t *testing.T) {
	hec := framework.NewHookExecutionConfig(t,
		&pkg.HookConfig{
			Metadata:     pkg.HookMetadata{Name: "openapi-with-existing-hook"},
			OnBeforeHelm: &pkg.OrderedConfig{Order: 1},
		},
		subfolder.HandlerHookValues,
		framework.WithOpenAPIDir(openapiDir),
		framework.WithInitialValues(`{
            "some": {
                "path": {
                    "to": {
                        "field": {
                            "someInt": 1,
                            "array":   [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
                        }
                    }
                }
            }
        }`),
	)
	hec.RunHook()
	require.NoError(t, hec.HookError())

	// Schema defaults made it through.
	assert.Equal(t, "Disabled", hec.ValuesGet("https.mode").String())
	assert.True(t, hec.ValuesGet("internal.golangVersions").Exists())

	// And the hook ran: it sets `.some.path.to.field.str` then removes
	// `.some.path.to.field`, so the resulting value has neither the
	// original nested object nor the str — the parent was removed last.
	assert.False(t, hec.ValuesGet("some.path.to.field").Exists(),
		"values hook removes the parent key as part of its handler")
}
