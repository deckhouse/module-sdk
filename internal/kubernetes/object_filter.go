package kubernetes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

type Snapshots map[string][]byte

var (
	ErrSnapshotIsNotFound = errors.New("snapshot is not found")
)

func (s Snapshots) EnrichStructByKey(key string, v any) error {
	snap, ok := s[key]
	if !ok {
		return ErrSnapshotIsNotFound
	}

	buf := bytes.NewBuffer(snap)
	err := json.NewDecoder(buf).Decode(v)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	return nil
}

type ObjectAndFilterResults map[string]*ObjectAndFilterResult

// ByNamespaceAndName implements sort.Interface for []ObjectAndFilterResult
// based on Namespace and Name of Object field.
type ByNamespaceAndName []ObjectAndFilterResult

func (a ByNamespaceAndName) Len() int      { return len(a) }
func (a ByNamespaceAndName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type ObjectAndFilterResult struct {
	Object       any `json:"object,omitempty"`
	FilterResult any `json:"filterResult,omitempty"`
}
