package utils

import (
	"fmt"

	k8syaml "sigs.k8s.io/yaml"
)

// NewValuesFromBytes loads values sections from maps in yaml or json format
func NewValuesFromBytes(data []byte) (map[string]any, error) {
	var values map[string]any

	err := k8syaml.Unmarshal(data, &values)
	if err != nil {
		return nil, fmt.Errorf("bad values data: %s\n%s", err, string(data))
	}

	return values, nil
}
