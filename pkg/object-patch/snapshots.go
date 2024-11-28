package objectpatch

import (
	"fmt"

	"github.com/deckhouse/module-sdk/pkg"
)

func UnmarshalToStruct[T any](s pkg.Snapshots, key string) ([]T, error) {
	snaps := s.Get(key)

	result := make([]T, 0, len(snaps))

	for _, snap := range snaps {
		obj := new(T)

		err := snap.UnmarhalTo(obj)
		if err != nil {
			return nil, fmt.Errorf("unmarshal to: %w", err)
		}

		result = append(result, *obj)
	}

	return result, nil
}
