package objectpatch

import "errors"

var (
	ErrSnapshotIsNotFound = errors.New("snapshot is not found")
)

func IgnoreSnapshotNotFound(err error) error {
	if errors.Is(err, ErrSnapshotIsNotFound) {
		return nil
	}

	return err
}
