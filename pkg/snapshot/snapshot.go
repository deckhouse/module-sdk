package snapshot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/deckhouse/module-sdk/pkg"
)

var _ pkg.Snapshot = (*Wrap)(nil)

type Wrap struct {
	Wrapped interface{}
}

func (w *Wrap) UnmarhalTo(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		// error replace with "not pointer"
		return errors.New(fmt.Sprintf("reflect.TypeOf(v): %s", reflect.TypeOf(v)))
	}

	// TODO: remove
	fmt.Println("input value: %v", rv)

	rw := reflect.ValueOf(w.Wrapped)
	if rw.Kind() != reflect.Pointer || rw.IsNil() {
		rv.Elem().Set(rw)

		return nil
	}

	fmt.Println("wrapped value: %v", rw)

	rv.Elem().Set(rw.Elem())

	return nil
}

func (w *Wrap) String() string {
	buf := bytes.NewBuffer([]byte{})
	_ = json.NewEncoder(buf).Encode(w)

	return buf.String()
}

// testing
// TODO: move to tests
type SomeStruct struct {
	String string
}

func Some() {
	wrapp := &Wrap{
		Wrapped: &SomeStruct{
			String: "INPUT STRING",
		},
	}

	ss := &SomeStruct{}

	err := wrapp.UnmarhalTo(ss)
	if err != nil {
		panic(err)
	}

	fmt.Println("output value: %v", ss)
}
