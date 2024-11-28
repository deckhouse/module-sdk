package objectpatch

import (
	"encoding/json"

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
