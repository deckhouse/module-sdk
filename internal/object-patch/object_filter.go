package objectpatch

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/deckhouse/module-sdk/pkg"
)

// inside json.RawMessage is array of json objects
type Snapshots map[string][]pkg.Snapshot

func (s Snapshots) Get(key string) []pkg.Snapshot {
	return s[key]
}

type Snapshot json.RawMessage

func (snap Snapshot) UnmarshalTo(v any) error {
	buf := bytes.NewBuffer(snap)

	err := json.NewDecoder(buf).Decode(v)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	return nil
}

func (snap Snapshot) String() string {
	return string(snap)
}
