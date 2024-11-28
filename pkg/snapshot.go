package pkg

import "encoding/json"

type Snapshots interface {
	Get(key string) []json.RawMessage
}
