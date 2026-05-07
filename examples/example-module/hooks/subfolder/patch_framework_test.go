package hookinfolder_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/framework"

	subfolder "example-module/subfolder"
)

// TestHandlerHookPatch_AppliesPatchesToFakeCluster runs the patch hook
// inside the framework. After RunHook() the framework replays every
// recorded patch operation against its fake cluster, so we can assert on
// the resulting cluster state directly.
func TestHandlerHookPatch_AppliesPatchesToFakeCluster(t *testing.T) {
	f := framework.HookExecutionConfigInit(t,
		&pkg.HookConfig{OnBeforeHelm: &pkg.OrderedConfig{Order: 1}},
		subfolder.HandlerHookPatch,
		`{}`, `{}`,
	)
	f.RunHook()

	require.NoError(t, f.HookError())

	// The hook calls Create + Delete on my-first-pod, so it should be gone.
	first := f.KubernetesResource("Pod", "default", "my-first-pod")
	assert.Nil(t, first, "my-first-pod should have been deleted after Create+Delete")

	// my-second-pod: CreateOrUpdate, then DeleteInBackground → also gone.
	second := f.KubernetesResource("Pod", "default", "my-second-pod")
	assert.Nil(t, second, "my-second-pod should have been deleted after Create+Delete")

	// my-third-pod: CreateIfNotExists, then DeleteNonCascading → gone too,
	// but not before being patched. Verify the hook recorded those calls.
	ops := f.PatchedOperations()
	require.Len(t, ops, 7)

	var sawMerge bool
	for _, op := range ops {
		if op.Type == framework.PatchTypeMergePatch && op.Name == "my-third-pod" {
			sawMerge = true
		}
	}
	assert.True(t, sawMerge, "expected a MergePatch on my-third-pod")
}
