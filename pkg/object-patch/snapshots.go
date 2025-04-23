package objectpatch

import (
	"fmt"
	"iter"

	"github.com/deckhouse/module-sdk/pkg"
)

func UnmarshalToStruct[T any](s pkg.Snapshots, key string) ([]T, error) {
	snaps := s.Get(key)

	result := make([]T, 0, len(snaps))

	for snap, err := range SnapshotIter[T](snaps) {
		if err != nil {
			return nil, err
		}

		result = append(result, snap)
	}

	return result, nil
}

func SnapshotIter[T any](snaps []pkg.Snapshot) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for _, snap := range snaps {
			obj := new(T)

			err := snap.UnmarshalTo(obj)
			if err != nil {
				err = fmt.Errorf("unmarshal to: %w", err)
			}

			if !yield(*obj, err) {
				break
			}
		}
	}
}
