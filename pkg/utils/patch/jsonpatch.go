package patch

// This is a copy of github.com/evanphx/json-patch v5.6.0+incompatible but that supports applying patches
// to internal objects. This is required to avoid sequential json marshal/unmarshal for each patch.
// (which make the behavior of the ApplyIgnoreNonExistentPaths method as much effective as ApplyStrict).

import (
	"encoding/json"

	lazynode "github.com/deckhouse/module-sdk/pkg/utils/lazy-node"
	"github.com/pkg/errors"
)

// Operation is a single JSON-Patch step, such as a single 'add' operation.
type Operation map[string]*json.RawMessage

// Kind reads the "op" field of the Operation.
func (o Operation) Kind() string {
	if obj, ok := o["op"]; ok && obj != nil {
		var op string

		err := json.Unmarshal(*obj, &op)
		if err != nil {
			return "unknown"
		}

		return op
	}

	return "unknown"
}

// Path reads the "path" field of the Operation.
func (o Operation) Path() (string, error) {
	if obj, ok := o["path"]; ok && obj != nil {
		var op string

		err := json.Unmarshal(*obj, &op)
		if err != nil {
			return "unknown", err
		}

		return op, nil
	}

	return "unknown", errors.Wrapf(ErrMissing, "operation missing path field")
}

// From reads the "from" field of the Operation.
func (o Operation) From() (string, error) {
	if obj, ok := o["from"]; ok && obj != nil {
		var op string

		err := json.Unmarshal(*obj, &op)
		if err != nil {
			return "unknown", err
		}

		return op, nil
	}

	return "unknown", errors.Wrapf(ErrMissing, "operation, missing from field")
}

func (o Operation) value() *lazynode.LazyNode {
	if obj, ok := o["value"]; ok {
		return lazynode.NewLazyNode(obj)
	}

	return nil
}

// ValueInterface decodes the operation value into an interface.
func (o Operation) ValueInterface() (interface{}, error) {
	if obj, ok := o["value"]; ok && obj != nil {
		var v interface{}

		err := json.Unmarshal(*obj, &v)
		if err != nil {
			return nil, err
		}

		return v, nil
	}

	return nil, errors.Wrapf(ErrMissing, "operation, missing value field")
}
