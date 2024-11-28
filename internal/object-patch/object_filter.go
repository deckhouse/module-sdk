package objectpatch

import (
	"encoding/json"
)

// inside json.RawMessage is array of json objects
type Snapshots map[string][]json.RawMessage

func (s Snapshots) Get(key string) []json.RawMessage {
	return s[key]
}
