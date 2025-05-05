package pkg

type Snapshots interface {
	Get(key string) []Snapshot
}

type Snapshot interface {
	UnmarshalTo(v any) error
	String() string
}
