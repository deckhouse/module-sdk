package operation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/deckhouse/module-sdk/pkg"
)

var _ pkg.MetricCollectorOption = (Option)(nil)

type Option func(o pkg.MetricCollectorOptionApplier)

func (opt Option) Apply(o pkg.MetricCollectorOptionApplier) {
	opt(o)
}

func WithGroup(group string) Option {
	return func(o pkg.MetricCollectorOptionApplier) {
		o.WithGroup(group)
	}
}

var _ pkg.MetricCollectorOptionApplier = (*Operation)(nil)

type Operation struct {
	Name    string            `json:"name"`
	Group   string            `json:"group,omitempty"`
	Action  string            `json:"action,omitempty"`
	Value   *float64          `json:"value,omitempty"`
	Buckets []float64         `json:"buckets,omitempty"`
	Labels  map[string]string `json:"labels"`
}

func (op Operation) WithGroup(group string) {
	op.Group = group //nolint: staticcheck
}

func (op Operation) Validate() error {
	var err error

	if op.Action == "" {
		err = errors.Join(err, fmt.Errorf("one of: 'action', 'set' or 'add' is required: %+v", op))
	}

	if op.Action != "set" && op.Action != "add" {
		if op.Group == "" && op.Action != "observe" {
			err = errors.Join(err, fmt.Errorf("unsupported action '%s': %+v", op.Action, op))
		}

		if op.Name == "" && op.Group != "" && op.Action != "expire" {
			err = errors.Join(err, fmt.Errorf("'name' is required when action is not 'expire': %+v", op))
		}
	}

	if op.Name == "" && op.Group == "" {
		err = errors.Join(err, fmt.Errorf("'name' is required: %+v", op))
	}

	if op.Action == "set" && op.Value == nil {
		err = errors.Join(err, fmt.Errorf("'value' is required for action 'set': %+v", op))
	}

	if op.Action == "add" && op.Value == nil {
		err = errors.Join(err, fmt.Errorf("'value' is required for action 'add': %+v", op))
	}

	if op.Action == "observe" && op.Value == nil {
		err = errors.Join(err, fmt.Errorf("'value' is required for action 'observe': %+v", op))
	}

	if op.Action == "observe" && op.Buckets == nil {
		err = errors.Join(err, fmt.Errorf("'buckets' is required for action 'observe': %+v", op))
	}

	return err
}

func MetricOperationsFromReader(r io.Reader) ([]Operation, error) {
	operations := make([]Operation, 0)

	dec := json.NewDecoder(r)
	for {
		var metricOperation Operation
		if err := dec.Decode(&metricOperation); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		operations = append(operations, metricOperation)
	}

	return operations, nil
}

func MetricOperationsFromBytes(data []byte) ([]Operation, error) {
	return MetricOperationsFromReader(bytes.NewReader(data))
}

func MetricOperationsFromFile(filePath string) ([]Operation, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %s", filePath, err)
	}

	if len(data) == 0 {
		return nil, nil
	}
	return MetricOperationsFromBytes(data)
}

func ValidateOperations(ops []Operation) error {
	var opsErrs error

	for _, op := range ops {
		err := op.Validate()
		if err != nil {
			opsErrs = errors.Join(opsErrs, err)
		}
	}

	return opsErrs
}
