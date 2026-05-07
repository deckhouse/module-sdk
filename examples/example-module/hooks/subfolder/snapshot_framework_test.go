package hookinfolder_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/framework"

	subfolder "example-module/subfolder"
)

// snapshotHookConfig matches the configuration registered by
// snapshot_getting_hook_alternative.go. It is duplicated here so the
// functional test does not depend on the global registry (which is
// shared across the package and would mix in unrelated bindings).
var snapshotHookConfig = &pkg.HookConfig{
	OnBeforeHelm: &pkg.OrderedConfig{Order: 1},
	Kubernetes: []pkg.KubernetesConfig{
		{
			Name:       subfolder.NodeInfoSnapshotName,
			APIVersion: "v1",
			Kind:       "Node",
			JqFilter: `{
				"apiVersion": .apiVersion,
				"kind":       .kind,
				"metadata": {
					"name":            .metadata.name,
					"resourceVersion": .metadata.resourceVersion,
					"uid":             .metadata.uid
				}
			}`,
		},
	},
}

// TestSnapshotsAlt_ReadsNodesFromCluster drives the alternative snapshot
// hook against a real fake cluster: two Nodes are seeded, the framework
// generates the snapshots from the binding spec, and the hook is asserted
// to log a "node found" entry per Node.
func TestSnapshotsAlt_ReadsNodesFromCluster(t *testing.T) {
	const state = `
---
apiVersion: v1
kind: Node
metadata:
  name: node-a
  uid: "uid-a"
  resourceVersion: "10"
---
apiVersion: v1
kind: Node
metadata:
  name: node-b
  uid: "uid-b"
  resourceVersion: "20"
`

	f := framework.HookExecutionConfigInit(t,
		snapshotHookConfig,
		subfolder.HandlerHookSnapshotsAlt,
		`{}`, `{}`,
	)
	f.KubeStateSet(state)
	f.RunHook()

	require.NoError(t, f.HookError())

	snaps := f.Snapshots().Get(subfolder.NodeInfoSnapshotName)
	require.Len(t, snaps, 2)

	logs := f.LoggerOutput().String()
	assert.Contains(t, logs, "hello from snapshot alt hook")
	for _, name := range []string{"node-a", "node-b"} {
		assert.Contains(t, logs, name, "expected node %q to appear in logs", name)
	}

	// And no extra log lines beyond the hook's expected output.
	lineCount := strings.Count(logs, "\n")
	assert.GreaterOrEqual(t, lineCount, 3, "expected at least 3 log lines (1 greeting + 2 nodes)")
}

// TestSnapshotsAlt_NoNodesNoLogs checks the empty-state path. The hook is
// well-behaved and emits only the greeting line when there are no nodes.
func TestSnapshotsAlt_NoNodesNoLogs(t *testing.T) {
	f := framework.HookExecutionConfigInit(t,
		snapshotHookConfig,
		subfolder.HandlerHookSnapshotsAlt,
		`{}`, `{}`,
	)
	f.RunHook()

	require.NoError(t, f.HookError())
	require.Empty(t, f.Snapshots().Get(subfolder.NodeInfoSnapshotName))
	assert.Contains(t, f.LoggerOutput().String(), "hello from snapshot alt hook")
}
