package helpers

import (
	"encoding/json"
	"fmt"

	k8syaml "sigs.k8s.io/yaml"

	"github.com/deckhouse/module-sdk/pkg"
)

// StaticSnapshots is an in-memory implementation of pkg.Snapshots backed
// by raw JSON payloads. It is the simplest possible Snapshots stub: every
// snapshot is a JSON document that UnmarshalTo will decode into the
// caller-supplied struct.
//
// Construct a StaticSnapshots with NewSnapshots:
//
//	snaps := helpers.NewSnapshots().
//	    Add("nodes", helpers.SnapshotJSON(`{"name":"n1"}`)).
//	    Add("nodes", helpers.SnapshotYAML(`name: n2`))
type StaticSnapshots map[string][]pkg.Snapshot

// NewSnapshots returns an empty StaticSnapshots.
func NewSnapshots() StaticSnapshots {
	return StaticSnapshots{}
}

// Get implements pkg.Snapshots.
func (s StaticSnapshots) Get(key string) []pkg.Snapshot {
	return s[key]
}

// Add appends one or more snapshots to the bucket identified by key.
// It returns the receiver for chaining.
func (s StaticSnapshots) Add(key string, snaps ...pkg.Snapshot) StaticSnapshots {
	s[key] = append(s[key], snaps...)
	return s
}

// Set replaces the bucket identified by key with the given snapshots.
// It returns the receiver for chaining.
func (s StaticSnapshots) Set(key string, snaps ...pkg.Snapshot) StaticSnapshots {
	s[key] = append([]pkg.Snapshot(nil), snaps...)
	return s
}

// jsonSnapshot is a pkg.Snapshot whose payload is a JSON document.
type jsonSnapshot struct {
	raw []byte
}

// String implements pkg.Snapshot.
func (j jsonSnapshot) String() string { return string(j.raw) }

// UnmarshalTo implements pkg.Snapshot.
func (j jsonSnapshot) UnmarshalTo(v any) error {
	if len(j.raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(j.raw, v); err != nil {
		return fmt.Errorf("snapshot unmarshal: %w", err)
	}
	return nil
}

// SnapshotJSON builds a pkg.Snapshot from a JSON string.
func SnapshotJSON(raw string) pkg.Snapshot {
	return jsonSnapshot{raw: []byte(raw)}
}

// SnapshotYAML builds a pkg.Snapshot from a YAML string. The YAML is
// converted to canonical JSON internally so that UnmarshalTo behaves the
// same as for SnapshotJSON.
func SnapshotYAML(raw string) pkg.Snapshot {
	data, err := k8syaml.YAMLToJSON([]byte(raw))
	if err != nil {
		panic(fmt.Errorf("helpers.SnapshotYAML: %w", err))
	}
	return jsonSnapshot{raw: data}
}

// SnapshotFromObject builds a pkg.Snapshot by JSON-encoding the given Go
// value. It panics if the value cannot be marshalled (which would be a
// programmer error in test code).
func SnapshotFromObject(v any) pkg.Snapshot {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Errorf("helpers.SnapshotFromObject: marshal: %w", err))
	}
	return jsonSnapshot{raw: data}
}

// SnapshotFromObjects is the bulk variant of SnapshotFromObject.
func SnapshotFromObjects[T any](items []T) []pkg.Snapshot {
	out := make([]pkg.Snapshot, 0, len(items))
	for i := range items {
		out = append(out, SnapshotFromObject(items[i]))
	}
	return out
}
