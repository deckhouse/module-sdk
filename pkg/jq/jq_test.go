package jq_test

import (
	"context"
	"testing"

	"github.com/deckhouse/module-sdk/pkg/jq"
	"github.com/stretchr/testify/assert"
)

func TestJqFilter(t *testing.T) {
	const jqFilter = `.metadata.name // "foobar"`

	query, err := jq.NewQuery(jqFilter)
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
