package testutils

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	. "github.com/onsi/gomega"
)

func Test_MergeValues(t *testing.T) {
	expectations := []struct {
		testName       string
		values1        Values
		values2        Values
		expectedValues Values
	}{
		{
			"simple",
			Values{"a": 1, "b": 2},
			Values{"b": 3, "c": 4},
			Values{"a": 1, "b": 3, "c": 4},
		},
		{
			"array",
			Values{"a": []any{1}},
			Values{"a": []any{2}},
			Values{"a": []any{2}},
		},
		{
			"map",
			Values{"a": map[string]any{"a": 1, "b": 2}},
			Values{"a": map[string]any{"b": 3, "c": 4}},
			Values{"a": map[string]any{"a": 1, "b": 3, "c": 4}},
		},
	}

	for _, expectation := range expectations {
		t.Run(expectation.testName, func(t *testing.T) {
			values := MergeValues(expectation.values1, expectation.values2)

			if !reflect.DeepEqual(expectation.expectedValues, values) {
				t.Errorf("\n[EXPECTED]: %#v\n[GOT]: %#v", expectation.expectedValues, values)
			}
		})
	}
}

func expectStringToEqual(str string, expected string) error {
	if str != expected {
		return fmt.Errorf("Expected '%s' string, got '%s'", expected, str)
	}
	return nil
}

func Test_Values_loaders(t *testing.T) {
	g := NewWithT(t)

	jsonInput := []byte(`{
"global": {
  "param1": "value1",
  "param2": "value2"},
"moduleOne": {
  "paramStr": "string",
  "paramNum": 123,
  "paramArr": ["H", "He", "Li"]}
}`)
	yamlInput := []byte(`
global:
  param1: value1
  param2: value2
moduleOne:
  paramStr: string
  paramNum: 123
  paramArr:
  - H
  - He
  - Li
`)
	mapInput := map[string]any{
		"global": map[string]string{
			"param1": "value1",
			"param2": "value2",
		},
		"moduleOne": map[string]any{
			"paramStr": "string",
			"paramNum": 123,
			"paramArr": []string{
				"H",
				"He",
				"Li",
			},
		},
	}

	expected := Values(map[string]any{
		"global": map[string]any{
			"param1": "value1",
			"param2": "value2",
		},
		"moduleOne": map[string]any{
			"paramArr": []any{
				"H",
				"He",
				"Li",
			},
			"paramNum": 123.0,
			"paramStr": "string",
		},
	})

	var values Values
	var err error

	values, err = NewValuesFromBytes(jsonInput)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(values).To(Equal(expected), "expected: %s\nvalues: %s\n", spew.Sdump(expected), spew.Sdump(values))

	values, err = NewValuesFromBytes(yamlInput)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(values).To(Equal(expected), "expected: %s\nvalues: %s\n", spew.Sdump(expected), spew.Sdump(values))

	values, err = NewValues(mapInput)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(values).To(Equal(expected))
}
