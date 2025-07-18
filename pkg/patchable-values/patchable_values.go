package patchablevalues

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/deckhouse/deckhouse/pkg/log"

	service "github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/utils"
)

var _ service.PatchableValuesCollector = (*PatchableValues)(nil)

type PatchableValues struct {
	values          *gjson.Result
	patchOperations []*utils.ValuesPatchOperation
}

func NewPatchableValues(values map[string]any) (*PatchableValues, error) {
	data, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	res := gjson.ParseBytes(data)

	return &PatchableValues{values: &res}, nil
}

// Get value from patchable. It could be null value
func (p *PatchableValues) Get(path string) gjson.Result {
	return p.values.Get(path)
}

// GetOk returns value and `exists` flag
func (p *PatchableValues) GetOk(path string) (gjson.Result, bool) {
	v := p.values.Get(path)
	if v.Exists() {
		return v, true
	}

	return v, false
}

// GetRaw get empty interface
func (p *PatchableValues) GetRaw(path string) any {
	return p.values.Get(path).Value()
}

// Exists checks whether a path exists
func (p *PatchableValues) Exists(path string) bool {
	return p.values.Get(path).Exists()
}

// ArrayCount counts the number of elements in a JSON array at a path
func (p *PatchableValues) ArrayCount(path string) (int, error) {
	v := p.values.Get(path)
	if !v.IsArray() {
		return 0, fmt.Errorf("value at %q path is not an array", path)
	}

	return len(v.Array()), nil
}

func (p *PatchableValues) Set(path string, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		// The struct returned from a Go hook expected to be marshalable in all cases.
		// TODO(nabokihms): return a meaningful error.
		log.Error("patch path",
			slog.String("path", path),
			log.Err(err))
		return
	}

	op := &utils.ValuesPatchOperation{
		Op:    "add",
		Path:  convertDotFilePathToSlashPath(path),
		Value: data,
	}

	p.patchOperations = append(p.patchOperations, op)
}

func (p *PatchableValues) Remove(path string) {
	if !p.Exists(path) {
		// return if path not exists
		return
	}

	op := &utils.ValuesPatchOperation{
		Op:   "remove",
		Path: convertDotFilePathToSlashPath(path),
	}

	p.patchOperations = append(p.patchOperations, op)
}

func (p *PatchableValues) GetPatches() []*utils.ValuesPatchOperation {
	return p.patchOperations
}

func (p *PatchableValues) WriteOutput(w io.Writer) error {
	if len(p.patchOperations) == 0 {
		return nil
	}

	err := json.NewEncoder(w).Encode(p.patchOperations)
	if err != nil {
		return err
	}

	return nil
}

func convertDotFilePathToSlashPath(dotPath string) string {
	return strings.ReplaceAll("/"+dotPath, ".", "/")
}
