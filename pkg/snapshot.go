package pkg

type Snapshots interface {
	Get(key string) []Snapshot
}

type Snapshot interface {
	UnmarhalTo(v any) error
}
