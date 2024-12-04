package objectpatch_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	objectpatch "github.com/deckhouse/module-sdk/internal/object-patch"
	pkgobjectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
)

func Test_Snapshots(t *testing.T) {
	t.Run("test strings snapshot unmarshal", func(t *testing.T) {
		snaps := make(objectpatch.Snapshots, 1)
		snaps["stub-key"] = append(snaps["stub-key"], objectpatch.Snapshot(`"one"`), objectpatch.Snapshot(`"two"`), objectpatch.Snapshot(`"three"`))

		strs, err := pkgobjectpatch.UnmarshalToStruct[string](snaps, "stub-key")
		assert.NoError(t, err)
		assert.Equal(t, strs, []string{"one", "two", "three"})
	})

	t.Run("test bools snapshot unmarshal", func(t *testing.T) {
		snaps := make(objectpatch.Snapshots, 1)
		snaps["stub-key"] = append(snaps["stub-key"], objectpatch.Snapshot(`true`), objectpatch.Snapshot(`false`), objectpatch.Snapshot(`true`))

		strs, err := pkgobjectpatch.UnmarshalToStruct[bool](snaps, "stub-key")
		assert.NoError(t, err)
		assert.Equal(t, strs, []bool{true, false, true})
	})

	t.Run("test uints snapshot unmarshal", func(t *testing.T) {
		snaps := make(objectpatch.Snapshots, 1)
		snaps["stub-key"] = append(snaps["stub-key"], objectpatch.Snapshot(`1`), objectpatch.Snapshot(`2`), objectpatch.Snapshot(`3`))

		strs, err := pkgobjectpatch.UnmarshalToStruct[uint](snaps, "stub-key")
		assert.NoError(t, err)
		assert.Equal(t, strs, []uint{1, 2, 3})
	})

	t.Run("test structs snapshot unmarshal", func(t *testing.T) {
		const (
			first = `
		{
			"metadata": {
				"name": "first-name"
			}
		}`
			second = `
		{
			"metadata": {
				"name": "second-name"
			}
		}`
			third = `
		{
			"metadata": {
				"name": "third-name"
			}
		}`
		)

		type TestMetadata struct {
			Name string `json:"name"`
		}

		type TestStruct struct {
			Metadata TestMetadata `json:"metadata"`
		}

		snaps := make(objectpatch.Snapshots, 1)
		snaps["stub-key"] = append(snaps["stub-key"], objectpatch.Snapshot(first), objectpatch.Snapshot(second), objectpatch.Snapshot(third))

		strs, err := pkgobjectpatch.UnmarshalToStruct[TestStruct](snaps, "stub-key")
		assert.NoError(t, err)
		assert.Equal(t, strs, []TestStruct{
			{Metadata: TestMetadata{Name: "first-name"}},
			{Metadata: TestMetadata{Name: "second-name"}},
			{Metadata: TestMetadata{Name: "third-name"}},
		})
	})
}
