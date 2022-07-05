package rshelper

import (
	"fmt"

	"arhat.dev/rs"
	"go.uber.org/multierr"

	"arhat.dev/pkg/envhelper"
	"arhat.dev/pkg/yamlhelper"
)

// EnvRenderingHandler expands rawData with environment variables with bash string replacement functions support
// Please refer to https://github.com/drone/envsubst
type EnvRenderingHandler struct {
	// Env used for expansion
	// Required if you do want to expand some env
	Env map[string]string

	AllowNotFound bool
}

func (h *EnvRenderingHandler) RenderYaml(
	_ string, rawData interface{},
) ([]byte, error) {
	rawData, err := rs.NormalizeRawData(rawData)
	if err != nil {
		return nil, err
	}

	bytesToExpand, err := yamlhelper.ToYamlBytes(rawData)
	if err != nil {
		return nil, fmt.Errorf("failed to get data bytes of input: %w", err)
	}

	var notFoundErr error
	s := envhelper.Expand(string(bytesToExpand), func(varName string, origin string) string {
		if h.Env == nil {
			if h.AllowNotFound {
				return ""
			}

			notFoundErr = multierr.Append(notFoundErr, fmt.Errorf("env %q not found", origin))
			return origin
		}

		v, ok := h.Env[varName]
		if !ok {
			if h.AllowNotFound {
				return ""
			}

			notFoundErr = multierr.Append(notFoundErr, fmt.Errorf("env %q not found", origin))
			return origin
		}

		return v
	})

	return []byte(s), notFoundErr
}
