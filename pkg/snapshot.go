package pkg

type Snapshots interface {
	UnmarshalToStruct(key string, v any) error
}
