package utils

import (
	"encoding/json"
	"fmt"

	"github.com/deckhouse/module-sdk/pkg/utils/patch"
)

type ValuesPatchType string

const (
	ConfigMapPatch    ValuesPatchType = "CONFIG_MAP_PATCH"
	MemoryValuesPatch ValuesPatchType = "MEMORY_VALUES_PATCH"
)

type ValuesPatchOperation struct {
	Op    string          `json:"op,omitempty"`
	Path  string          `json:"path,omitempty"`
	Value json.RawMessage `json:"value,omitempty"`
}

func (op *ValuesPatchOperation) ToString() string {
	data, err := json.Marshal(op.Value)
	if err != nil {
		// This should not happen, because ValuesPatchOperation is created with Unmarshal!
		return fmt.Sprintf("{\"op\":\"%s\", \"path\":\"%s\", \"value-error\": \"%s\" }", op.Op, op.Path, err)
	}
	return string(data)
}

// ToJSONPatch returns a jsonpatch.Patch with one operation.
func (op *ValuesPatchOperation) ToJSONPatch() (patch.Patch, error) {
	opBytes, err := json.Marshal([]*ValuesPatchOperation{op})
	if err != nil {
		return nil, err
	}
	return DecodePatch(opBytes)
}

// DecodePatch decodes the passed JSON document as an RFC 6902 patch.
func DecodePatch(buf []byte) (patch.Patch, error) {
	var p patch.Patch

	err := json.Unmarshal(buf, &p)
	if err != nil {
		return nil, err
	}

	return p, nil
}
