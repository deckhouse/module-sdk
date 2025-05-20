package pkg

type Snapshots interface {
	Get(key string) []Snapshot
}

type Snapshot interface {
	UnmarshalTo(v any) error
	// returns pure form of object
	// can contains special symbols
	// to receive string values - use UnmarshalTo method
	String() string
}
