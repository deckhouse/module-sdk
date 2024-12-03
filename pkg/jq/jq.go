package jq

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/itchyny/gojq"
)

type Result struct {
	data any
}

func (res *Result) String() string {
	return gojq.Preview(res.data)
}

func (res *Result) TypeOf() string {
	return gojq.TypeOf(res.data)
}

type Query struct {
	payload string
	query   *gojq.Query
	code    *gojq.Code
}

func NewJQ(query string) (*Query, error) {
	parsedQuery, err := gojq.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	code, err := gojq.Compile(parsedQuery)
	if err != nil {
		return nil, fmt.Errorf("compile: %w", err)
	}

	return &Query{
		payload: query,
		query:   parsedQuery,
		code:    code,
	}, nil
}

func (jq *Query) FilterObject(ctx context.Context, v any) ([]Result, error) {
	buf := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(buf).Encode(v)
	if err != nil {
		return nil, fmt.Errorf("encode object: %w", err)
	}

	input := make(map[string]any, 1)
	err = json.NewDecoder(buf).Decode(&input)
	if err != nil {
		return nil, fmt.Errorf("decode object: %w", err)
	}

	var errs error
	result := make([]Result, 0, 1)
	iter := jq.code.RunWithContext(ctx, input)

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := v.(error); ok {
			if err, ok := err.(*gojq.HaltError); ok && err.Value() == nil {
				break
			}

			errs = errors.Join(errs, err)
		}

		result = append(result, Result{data: v})
	}

	return result, errs
}

func Validate(query string) error {
	_, err := gojq.Parse(query)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	return nil
}
