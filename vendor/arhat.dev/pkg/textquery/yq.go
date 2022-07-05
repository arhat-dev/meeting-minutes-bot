package textquery

import (
	"io"
	"strings"

	"arhat.dev/pkg/stringhelper"
	"gopkg.in/yaml.v3"
)

// YQ runs jq over yaml data
func YQ[B ~byte, T stringhelper.String[B]](query string, data T) (string, error) {
	var (
		rd strings.Reader
		sb strings.Builder
	)

	rd.Reset(stringhelper.Convert[string, B](data))

	err := Query(
		query,
		nil,
		NewYAMLIterFunc(&rd),
		CreateResultToTextHandleFuncForJsonOrYaml(&sb, yaml.Marshal),
	)

	return sb.String(), err
}

func NewYAMLIterFunc(r io.Reader) func() (any, bool) {
	dec := yaml.NewDecoder(r)

	return func() (any, bool) {
		var data any
		err := dec.Decode(&data)
		if err != nil {
			// return plain text on unexpected error
			return nil, false
		}

		return data, true
	}
}
