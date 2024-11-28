package objectpatch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/deckhouse/module-sdk/pkg"
)

var (
	ErrSnapshotIsNotFound = errors.New("snapshot is not found")
)

func IgnoreSnapshotIsNotFound(err error) error {
	if errors.Is(err, ErrSnapshotIsNotFound) {
		return nil
	}

	return err
}

func UnmarshalToStruct[T any](s pkg.Snapshots, key string) ([]T, error) {
	snaps := s.Get(key)

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
