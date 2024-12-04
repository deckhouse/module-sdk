package jq

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJqFilter(t *testing.T) {
	const jqFilter = `.metadata.name // "foobar"`
	query, err := NewQuery(jqFilter)
	if err != nil {
		panic(err)
	}

	testCases := map[string]string{ // source: result
		`{"metadata":{"name":"stub"}}`: `"stub"`,
		`{"metadata":{}}`:              `"foobar"`,
	}

	for source, result := range testCases {
		ress, err := query.FilterStringObject(context.TODO(), source)
		if err != nil {
			panic(err)
		}

		if ress.String() != result {
			assert.Equal(t, ress.String(), result)
		}
	}
}
