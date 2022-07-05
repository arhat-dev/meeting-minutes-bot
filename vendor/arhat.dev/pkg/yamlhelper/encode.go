package yamlhelper

import (
	"gopkg.in/yaml.v3"
)

func ToYamlBytes(in interface{}) ([]byte, error) {
	switch t := in.(type) {
	case string:
		return []byte(t), nil
	case []byte:
		return t, nil
	default:
	}

	return yaml.Marshal(in)
}
