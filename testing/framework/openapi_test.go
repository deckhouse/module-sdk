package framework_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/framework"
)

// openapiStub points at the self-contained OpenAPI fixtures used by the
// framework's own tests. Tests must not depend on schemas defined in the
// example modules — those evolve independently as documentation.
const openapiStub = "openapi_stub"

// TestSchemaDefaults_FromConfigValues verifies that SchemaDefaults walks
// an OpenAPI schema and produces a values map containing every `default:`
// value declared in it (recursively).
func TestSchemaDefaults_FromConfigValues(t *testing.T) {
	schema, err := framework.LoadOpenAPISchema(filepath.Join(openapiStub, "config-values.yaml"))
	require.NoError(t, err)

	defaults := framework.SchemaDefaults(schema)

	assert.EqualValues(t, 1, defaults["replicas"])

	https, _ := defaults["https"].(map[string]any)
	require.NotNil(t, https)
	assert.Equal(t, "Disabled", https["mode"])

	cm, _ := https["certManager"].(map[string]any)
	require.NotNil(t, cm)
	assert.Equal(t, "letsencrypt", cm["clusterIssuerName"])

	// `customCertificate` declares `default: {}` AND a `secretName`
	// sub-property with its own default. Our defaulting algorithm must
	// merge sub-property defaults into the explicit default object.
	custom, _ := https["customCertificate"].(map[string]any)
	require.NotNil(t, custom)
	assert.Equal(t, "false", custom["secretName"])

	// Top-level array default must be preserved as-is.
	tags, ok := defaults["tags"].([]any)
	require.True(t, ok, "expected tags to be a slice, got %T", defaults["tags"])
	assert.Empty(t, tags)
}

// TestSchemaDefaults_XExtend verifies that values.yaml correctly inherits
// properties from config-values.yaml when it declares `x-extend`.
func TestSchemaDefaults_XExtend(t *testing.T) {
	schema, err := framework.LoadOpenAPISchema(filepath.Join(openapiStub, "values.yaml"))
	require.NoError(t, err)

	defaults := framework.SchemaDefaults(schema)

	internal, _ := defaults["internal"].(map[string]any)
	require.NotNil(t, internal, "values.yaml should produce default for internal")
	versions, _ := internal["golangVersions"].([]any)
	assert.Empty(t, versions)

	assert.EqualValues(t, 1, defaults["replicas"], "should inherit replicas default from config-values.yaml")

	https, _ := defaults["https"].(map[string]any)
	require.NotNil(t, https, "should inherit https tree from config-values.yaml")
	assert.Equal(t, "Disabled", https["mode"])

	// `registry` has no default anywhere in the schema and must not appear.
	_, hasRegistry := defaults["registry"]
	assert.False(t, hasRegistry, "registry has no defaults and must not be synthesised")
}

// TestMergeValues verifies the deep-merge semantics: objects are merged
// property-by-property, scalars and arrays are replaced.
func TestMergeValues(t *testing.T) {
	base := map[string]any{
		"replicas": 1,
		"https": map[string]any{
			"mode": "Disabled",
			"certManager": map[string]any{
				"clusterIssuerName": "letsencrypt",
			},
		},
		"tags": []any{"a", "b"},
	}
	override := map[string]any{
		"replicas": 5,
		"https": map[string]any{
			"mode": "CertManager",
		},
		"tags": []any{"c"},
	}

	out := framework.MergeValues(base, override)

	assert.EqualValues(t, 5, out["replicas"])

	https, _ := out["https"].(map[string]any)
	require.NotNil(t, https)
	assert.Equal(t, "CertManager", https["mode"])
	cm, _ := https["certManager"].(map[string]any)
	require.NotNil(t, cm, "untouched nested map should survive merge")
	assert.Equal(t, "letsencrypt", cm["clusterIssuerName"])

	assert.Equal(t, []any{"c"}, out["tags"], "arrays must be replaced, not concatenated")

	// base should not be mutated.
	assert.EqualValues(t, 1, base["replicas"])
	baseHTTPS, _ := base["https"].(map[string]any)
	assert.Equal(t, "Disabled", baseHTTPS["mode"])
}

// TestWithOpenAPIDir_ProducesDefaults verifies the framework option:
// pointing at the openapi_stub/ directory must populate the values store
// with all defaults declared in the schema, with user-supplied initial
// values overriding them.
func TestWithOpenAPIDir_ProducesDefaults(t *testing.T) {
	cfg := &pkg.HookConfig{Metadata: pkg.HookMetadata{Name: "openapi-defaults"}}
	handler := func(_ context.Context, _ *pkg.HookInput) error { return nil }

	hec := framework.NewHookExecutionConfig(t, cfg, handler,
		framework.WithOpenAPIDir(openapiStub),
		framework.WithInitialValues(`{"https": {"mode": "CertManager"}}`),
		framework.WithInitialConfigValues(`{"replicas": 7}`),
	)
	hec.RunHook()
	require.NoError(t, hec.HookError())

	assert.EqualValues(t, 7, hec.ConfigValuesGet("replicas").Int(),
		"user override must win for config values")

	assert.Equal(t, "Disabled", hec.ConfigValuesGet("https.mode").String(),
		"untouched config-values defaults must remain")
	assert.Equal(t, "letsencrypt", hec.ConfigValuesGet("https.certManager.clusterIssuerName").String())

	assert.Equal(t, "CertManager", hec.ValuesGet("https.mode").String(),
		"user override on values must win")
	assert.Equal(t, "letsencrypt", hec.ValuesGet("https.certManager.clusterIssuerName").String(),
		"values must inherit defaults from config-values via x-extend")

	assert.True(t, hec.ValuesGet("internal.golangVersions").Exists(),
		"values.yaml-only defaults must also be present")
}

// TestWithOpenAPIDir_IgnoresMissingFiles verifies that WithOpenAPIDir is
// permissive: pointing at a directory that has neither values.yaml nor
// config-values.yaml must succeed and behave as if the option had not
// been passed.
func TestWithOpenAPIDir_IgnoresMissingFiles(t *testing.T) {
	emptyDir := t.TempDir()

	cfg := &pkg.HookConfig{Metadata: pkg.HookMetadata{Name: "openapi-empty"}}
	handler := func(_ context.Context, _ *pkg.HookInput) error { return nil }

	hec := framework.NewHookExecutionConfig(t, cfg, handler,
		framework.WithOpenAPIDir(emptyDir),
		framework.WithInitialValues(`{"foo": "bar"}`),
	)
	hec.RunHook()
	require.NoError(t, hec.HookError())
	assert.Equal(t, "bar", hec.ValuesGet("foo").String())
}

// TestWithValuesSchema_AppliesDefaults verifies the targeted variant:
// only the values schema is loaded; config-values are left untouched.
func TestWithValuesSchema_AppliesDefaults(t *testing.T) {
	cfg := &pkg.HookConfig{Metadata: pkg.HookMetadata{Name: "openapi-values-only"}}
	handler := func(_ context.Context, _ *pkg.HookInput) error { return nil }

	hec := framework.NewHookExecutionConfig(t, cfg, handler,
		framework.WithValuesSchema(filepath.Join(openapiStub, "values.yaml")),
	)
	hec.RunHook()
	require.NoError(t, hec.HookError())

	assert.Equal(t, "Disabled", hec.ValuesGet("https.mode").String())
	// config values are untouched.
	assert.False(t, hec.ConfigValuesGet("https.mode").Exists())
}

// TestWithValuesSchema_MissingFile fails fast.
func TestWithValuesSchema_MissingFile(t *testing.T) {
	cfg := &pkg.HookConfig{Metadata: pkg.HookMetadata{Name: "openapi-missing"}}
	handler := func(_ context.Context, _ *pkg.HookInput) error { return nil }

	tt := &captureT{TB: t}
	defer func() { _ = recover() }()

	framework.NewHookExecutionConfig(tt, cfg, handler,
		framework.WithValuesSchema("/this/path/does/not/exist.yaml"),
	)

	assert.True(t, tt.failed, "missing schema file should fail the test")
}

// captureT is a testing.TB that records Fatalf rather than killing the
// whole goroutine — used to assert that the framework reports errors.
type captureT struct {
	testing.TB
	failed bool
}

func (c *captureT) Fatalf(format string, args ...any) {
	c.failed = true
	// Don't actually call FailNow on the parent; we want the test that
	// uses captureT to keep running and assert on c.failed.
	panic("captureT: " + format)
}

func (c *captureT) Errorf(format string, args ...any) { c.failed = true }
