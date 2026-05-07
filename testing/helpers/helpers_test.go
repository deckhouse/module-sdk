package helpers_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/pkg"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	"github.com/deckhouse/module-sdk/testing/helpers"
)

func TestStaticSnapshots_AddAndUnmarshal(t *testing.T) {
	type Item struct {
		Name string `json:"name"`
	}

	snaps := helpers.NewSnapshots().
		Add("things",
			helpers.SnapshotJSON(`{"name":"a"}`),
			helpers.SnapshotYAML("name: b"),
		).
		Add("things", helpers.SnapshotFromObject(Item{Name: "c"}))

	got, err := objectpatch.UnmarshalToStruct[Item](snaps, "things")
	require.NoError(t, err)
	assert.Equal(t, []Item{{Name: "a"}, {Name: "b"}, {Name: "c"}}, got)
}

func TestStaticSnapshots_FromObjects(t *testing.T) {
	type Item struct{ Name string }

	source := []Item{{Name: "x"}, {Name: "y"}}
	snaps := helpers.NewSnapshots().Set("k", helpers.SnapshotFromObjects(source)...)

	got := snaps.Get("k")
	require.Len(t, got, 2)
	assert.JSONEq(t, `{"Name":"x"}`, got[0].String())
}

func TestStaticSnapshots_GetUnknownKey(t *testing.T) {
	snaps := helpers.NewSnapshots()
	assert.Empty(t, snaps.Get("missing"))
}

func TestNewValuesFromJSON_RoundTrip(t *testing.T) {
	v := helpers.NewValuesFromJSON(`{"foo":{"bar":"baz"}}`)
	assert.Equal(t, "baz", v.Get("foo.bar").String())

	v.Set("foo.qux", "quux")
	assert.Len(t, v.GetPatches(), 1)

	op := v.GetPatches()[0]
	assert.Equal(t, "add", op.Op)
	assert.Equal(t, "/foo/qux", op.Path)
}

func TestNewValuesFromYAML_Empty(t *testing.T) {
	v := helpers.NewValuesFromYAML("")
	assert.False(t, v.Exists("anything"))
}

func TestRecordingPatchCollector_RecordsAllOps(t *testing.T) {
	pc := helpers.NewRecordingPatchCollector()

	pc.Create(map[string]any{"kind": "Pod", "metadata": map[string]any{"name": "a"}})
	pc.CreateOrUpdate(map[string]any{"kind": "Pod"})
	pc.CreateIfNotExists(map[string]any{"kind": "Pod"})
	pc.Delete("v1", "Pod", "default", "to-delete")
	pc.DeleteInBackground("v1", "Pod", "default", "bg")
	pc.DeleteNonCascading("v1", "Pod", "default", "nc")
	pc.PatchWithMerge(map[string]any{"x": 1}, "v1", "Pod", "default", "p", objectpatch.WithIgnoreMissingObject(true))
	pc.PatchWithJSON([]map[string]any{{"op": "add", "path": "/x", "value": 1}}, "v1", "Pod", "default", "p")
	pc.PatchWithJQ(`.x = 1`, "v1", "Pod", "default", "p")

	recorded := pc.Recorded()
	require.Len(t, recorded, 9)

	gotOps := make([]string, 0, len(recorded))
	for _, op := range recorded {
		gotOps = append(gotOps, op.Op)
	}
	assert.Equal(t, []string{
		"Create", "CreateOrUpdate", "CreateIfNotExists",
		"Delete", "DeleteInBackground", "DeleteNonCascading",
		"MergePatch", "JSONPatch", "JQFilter",
	}, gotOps)
}

func TestRecordingPatchCollector_Filter(t *testing.T) {
	pc := helpers.NewRecordingPatchCollector()
	pc.Create(map[string]any{"kind": "Pod"})
	pc.Delete("v1", "Pod", "ns", "a")
	pc.DeleteInBackground("v1", "Pod", "ns", "b")

	deletes := pc.Filter("Delete", "DeleteInBackground")
	require.Len(t, deletes, 2)
	assert.Equal(t, "a", deletes[0].Name)
	assert.Equal(t, "b", deletes[1].Name)
}

func TestRecordingPatchCollector_OperationsInterface(t *testing.T) {
	pc := helpers.NewRecordingPatchCollector()
	pc.Delete("v1", "Pod", "ns", "x")

	ops := pc.Operations()
	require.Len(t, ops, 1)
	assert.Equal(t, "Delete", ops[0].Description())

	ops[0].SetObjectPrefix("test")
	require.Len(t, pc.Recorded(), 1)
	assert.Equal(t, "test-x", pc.Recorded()[0].Name)
}

func TestInputBuilder_DefaultsAreUsable(t *testing.T) {
	in := helpers.NewInputBuilder(t).Build()

	require.NotNil(t, in.Snapshots)
	require.NotNil(t, in.Values)
	require.NotNil(t, in.ConfigValues)
	require.NotNil(t, in.PatchCollector)
	require.NotNil(t, in.MetricsCollector)
	require.NotNil(t, in.Logger)

	in.Values.Set("a.b", "c")
	require.Len(t, in.Values.GetPatches(), 1)
	in.PatchCollector.Create(map[string]any{"kind": "Pod"})
	require.Len(t, in.PatchCollector.Operations(), 1)
}

func TestInputBuilder_FluentChain(t *testing.T) {
	type Item struct {
		Name string `json:"name"`
	}

	b := helpers.NewInputBuilder(t).
		WithSnapshot("items",
			helpers.SnapshotJSON(`{"name":"first"}`),
			helpers.SnapshotJSON(`{"name":"second"}`),
		).
		WithValuesJSON(`{"my":{"val":1}}`).
		WithConfigValuesYAML("module:\n  enabled: true\n").
		WithCapturedLogger().
		WithRecordingPatchCollector()

	in := b.Build()

	require.NotNil(t, b.LogBuffer())
	require.NotNil(t, b.RecordingPatchCollector())

	items, err := unmarshalSnapshots[Item](in.Snapshots, "items")
	require.NoError(t, err)
	assert.Equal(t, []Item{{Name: "first"}, {Name: "second"}}, items)

	assert.Equal(t, int64(1), in.Values.Get("my.val").Int())
	assert.True(t, in.ConfigValues.Get("module.enabled").Bool())

	in.Logger.Info("hello")
	assert.Contains(t, b.LogBuffer().String(), "hello")

	in.PatchCollector.Create(map[string]any{"kind": "Pod"})
	require.Len(t, b.RecordingPatchCollector().Recorded(), 1)
}

func TestJQRunOnString_AndObject(t *testing.T) {
	const filter = `{name: .metadata.name, count: (.spec.replicas // 0)}`
	const input = `{"metadata":{"name":"deploy"},"spec":{"replicas":3}}`

	type result struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	t.Run("string input", func(t *testing.T) {
		var got result
		require.NoError(t, helpers.JQRunOnString(context.Background(), filter, input, &got))
		assert.Equal(t, result{Name: "deploy", Count: 3}, got)
	})

	t.Run("object input", func(t *testing.T) {
		var got result
		var asMap map[string]any
		require.NoError(t, json.Unmarshal([]byte(input), &asMap))
		require.NoError(t, helpers.JQRunOnObject(context.Background(), filter, asMap, &got))
		assert.Equal(t, result{Name: "deploy", Count: 3}, got)
	})

	t.Run("nil target rejected", func(t *testing.T) {
		err := helpers.JQRunOnString(context.Background(), filter, input, nil)
		require.Error(t, err)
	})
}

// unmarshalSnapshots is a tiny test helper to avoid importing object-patch
// in every assertion above.
func unmarshalSnapshots[T any](s pkg.Snapshots, key string) ([]T, error) {
	out := []T{}
	for _, snap := range s.Get(key) {
		var v T
		if err := snap.UnmarshalTo(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}
