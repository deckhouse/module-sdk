package objectpatch

import (
	"bytes"
	"encoding/json"
	"fmt"

	pkgobjectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
)

// inside json.RawMessage is array of json objects
type Snapshots map[string]json.RawMessage

func (s Snapshots) UnmarshalToStruct(key string, v any) error {
	snap, ok := s[key]
	if !ok {
		return pkgobjectpatch.ErrSnapshotIsNotFound
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
