package utils

import (
	"errors"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/deckhouse/deckhouse/pkg/log"
	k8syaml "sigs.k8s.io/yaml"
)

const (
	GlobalValuesKey = "global"
)

// Values stores values for modules or hooks by name.
type Values map[string]interface{}

func GetValues() Values {
	values, err := LoadValuesFromFile(os.Getenv("VALUES_PATH"))
	if err != nil {
		return nil
	}
	return values
}

func GetConfigValues() Values {
	values, err := LoadValuesFromFile(os.Getenv("CONFIG_VALUES_PATH"))
	if err != nil {
		return nil
	}
	return values
}

func LoadValuesFromFile(valuesFilePath string) (Values, error) {
	valuesYaml, err := os.ReadFile(valuesFilePath)
	if err != nil && os.IsNotExist(err) {
		log.Debugf("No values file '%s': %v", valuesFilePath, err)
		return nil, nil
	}
	if err != nil {
		return nil, errors.Join(err, errors.New("load values file '"+valuesFilePath+"'"))
	}

	values, err := NewValuesFromBytes(valuesYaml)
	if err != nil {
		return nil, err
	}

	return values, nil
}

// NewValuesFromBytes loads values sections from maps in yaml or json format
func NewValuesFromBytes(data []byte) (Values, error) {
	var values map[string]interface{}

	err := k8syaml.Unmarshal(data, &values)
	if err != nil {
		return nil, fmt.Errorf("bad values data: %s\n%s", err, string(data))
	}

	return values, nil
}

// NewValues load all sections from input data and makes sure that input map
// can be marshaled to yaml and that yaml is compatible with json.
func NewValues(data map[string]interface{}) (Values, error) {
	yamlDoc, err := k8syaml.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("data is not compatible with JSON and YAML: %s, data:\n%s", err, spew.Sdump(data))
	}

	var values Values
	if err := k8syaml.Unmarshal(yamlDoc, &values); err != nil {
		return nil, fmt.Errorf("convert data YAML to values: %s, data:\n%s", err, spew.Sdump(data))
	}

	return values, nil
}

func MergeValues(values ...Values) Values {
	res := make(Values)

	for _, v := range values {
		res = MergeMap(res, v)
	}

	return res
}
