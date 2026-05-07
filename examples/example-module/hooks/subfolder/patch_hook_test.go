package hookinfolder_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"github.com/deckhouse/module-sdk/testing/helpers"

	subfolder "example-module/subfolder"
)

// TestHandlerHookPatch_RecordsExpectedOperations verifies that the patch
// hook issues the full sequence of Create/Delete/Patch operations against
// its PatchCollector.
//
// The RecordingPatchCollector lets us assert on each call without the
// minimock boilerplate the original Ginkgo test required.
func TestHandlerHookPatch_RecordsExpectedOperations(t *testing.T) {
	b := helpers.NewInputBuilder(t).
		WithRecordingPatchCollector().
		WithCapturedLogger()

	require.NoError(t, subfolder.HandlerHookPatch(context.Background(), b.Build()))

	ops := b.RecordingPatchCollector().Recorded()
	require.Len(t, ops, 7, "expected 3 creates, 3 deletes and 1 merge patch")

	// Creates
	assert.Equal(t, "Create", ops[0].Op)
	assert.Equal(t, "my-first-pod", ops[0].Object.(*corev1.Pod).Name)

	assert.Equal(t, "CreateOrUpdate", ops[1].Op)
	assert.Equal(t, "my-second-pod", ops[1].Object.(*corev1.Pod).Name)

	assert.Equal(t, "CreateIfNotExists", ops[2].Op)
	assert.Equal(t, "my-third-pod", ops[2].Object.(*corev1.Pod).Name)

	// Deletes
	for i, expected := range []struct {
		op   string
		name string
	}{
		{"Delete", "my-first-pod"},
		{"DeleteInBackground", "my-second-pod"},
		{"DeleteNonCascading", "my-third-pod"},
	} {
		op := ops[3+i]
		assert.Equal(t, expected.op, op.Op)
		assert.Equal(t, "v1", op.APIVersion)
		assert.Equal(t, "Pod", op.Kind)
		assert.Equal(t, "default", op.Namespace)
		assert.Equal(t, expected.name, op.Name)
	}

	// Merge patch with options
	mp := ops[6]
	assert.Equal(t, "MergePatch", mp.Op)
	assert.Equal(t, "my-third-pod", mp.Name)
	assert.Equal(t, map[string]any{"/status": "newStatus"}, mp.Patch)
	assert.Len(t, mp.Options, 2, "WithSubresource + WithIgnoreMissingObject expected")

	// Logger asserts
	logs := strings.Split(b.LogBuffer().String(), "\n")
	assert.Contains(t, logs[0], `"msg":"hello from patch hook"`)
}
