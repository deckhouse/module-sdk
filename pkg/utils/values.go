package utils //nolint: revive

import (
	"encoding/json"
	"fmt"

	"github.com/ettle/strcase"
	"gopkg.in/yaml.v3"
	k8syaml "sigs.k8s.io/yaml"
)

// ModuleNameToValuesKey returns camelCased name from kebab-cased (very-simple-module become verySimpleModule)
func ModuleNameToValuesKey(moduleName string) string {
	return strcase.ToCamel(moduleName)
}

// ModuleNameFromValuesKey returns kebab-cased name from camelCased (verySimpleModule become very-simple-module)
func ModuleNameFromValuesKey(moduleValuesKey string) string {
	return strcase.ToKebab(moduleValuesKey)
}

// NewValuesFromBytes loads values sections from maps in yaml or json format
func NewValuesFromBytes(data []byte) (map[string]any, error) {
	var values map[string]any

	err := k8syaml.Unmarshal(data, &values)
	if err != nil {
		return nil, fmt.Errorf("bad values data: %s\n%s", err, string(data))
	}

	return values, nil
}

func YamlBytes(v map[string]any) ([]byte, error) {
	return AsBytes(v, "yaml")
}

func AsBytes(v map[string]any, format string) ([]byte, error) {
	switch format {
	case "json":
		return json.Marshal(v)
	case "yaml":
		fallthrough
	default:
		return yaml.Marshal(v)
	}
}
