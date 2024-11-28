package objectpatch

import (
	"bytes"
	"encoding/json"
	"fmt"

	pkgobjectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
)

// inside json.RawMessage is array of json objects
type Snapshots map[string][]json.RawMessage

func (s Snapshots) GetObjectByKey(key string) ([]json.RawMessage, error) {
	snaps, ok := s[key]
	if !ok {
		return nil, pkgobjectpatch.ErrSnapshotIsNotFound
	}

	return snaps, nil
}

func UnmarshalToStruct[T any](s Snapshots, key string) ([]T, error) {
	snaps, ok := s[key]
	if !ok {
		return nil, pkgobjectpatch.ErrSnapshotIsNotFound
	}

	result := make([]T, 0, len(snaps))

	for _, snap := range snaps {
		obj := new(T)

		buf := bytes.NewBuffer(snap)

		err := json.NewDecoder(buf).Decode(obj)
		if err != nil {
			return nil, fmt.Errorf("decode: %w", err)
		}

		result = append(result, *obj)
	}

	return result, nil
}
