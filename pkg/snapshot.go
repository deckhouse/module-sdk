package pkg

import "encoding/json"

type Snapshots interface {
	GetObjectByKey(key string) ([]json.RawMessage, error)
}
