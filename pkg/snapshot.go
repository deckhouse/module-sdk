package pkg

type Snapshots interface {
	EnrichStructByKey(key string, v any) error
}
