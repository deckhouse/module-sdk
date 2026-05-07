package helpers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/deckhouse/module-sdk/pkg/jq"
)

// JQRunOnString applies the given JQ filter to a JSON string and decodes
// the result into target. target must be a non-nil pointer.
//
// This is a small convenience over manually constructing a jq.Query — it
// keeps unit tests focused on the assertions instead of the boilerplate.
func JQRunOnString(ctx context.Context, filter, jsonInput string, target any) error {
	if target == nil {
		return fmt.Errorf("helpers.JQRunOnString: target must be non-nil")
	}

	q, err := jq.NewQuery(filter)
	if err != nil {
		return fmt.Errorf("compile jq: %w", err)
	}

	res, err := q.FilterStringObject(ctx, jsonInput)
	if err != nil {
		return fmt.Errorf("apply jq: %w", err)
	}

	if err := json.Unmarshal([]byte(res.String()), target); err != nil {
		return fmt.Errorf("decode jq result: %w", err)
	}
	return nil
}

// JQRunOnObject applies the given JQ filter to a Go value (which must be
// JSON-serialisable) and decodes the result into target.
func JQRunOnObject(ctx context.Context, filter string, input, target any) error {
	if target == nil {
		return fmt.Errorf("helpers.JQRunOnObject: target must be non-nil")
	}

	q, err := jq.NewQuery(filter)
	if err != nil {
		return fmt.Errorf("compile jq: %w", err)
	}

	res, err := q.FilterObject(ctx, input)
	if err != nil {
		return fmt.Errorf("apply jq: %w", err)
	}

	if err := json.Unmarshal([]byte(res.String()), target); err != nil {
		return fmt.Errorf("decode jq result: %w", err)
	}
	return nil
}
