package interpreter

import (
	"fmt"
	"text/template"

	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/pkg/textquery"
	"github.com/Masterminds/sprig/v3"
)

// nolint:revive
const (
	Name = "interpreter"
)

func init() {
	publisher.Register(Name, func() publisher.Config { return &Config{} })
}

type Config struct {
	Bin  string   `json:"bin" yaml:"bin"`
	Args []string `json:"args" yaml:"args"`
}

func (c *Config) Create() (publisher.Interface, publisher.User, error) {
	var argTpls []*template.Template
	for _, arg := range c.Args {
		tpl, err := template.New("").
			Funcs(sprig.TxtFuncMap()).
			Funcs(map[string]interface{}{
				"jq":      textquery.JQ[byte, string],
				"jqBytes": textquery.JQ[byte, []byte],
			}).Parse(arg)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse arg template %q: %w", arg, err)
		}

		argTpls = append(argTpls, tpl)
	}

	return &Driver{
		bin:     c.Bin,
		argTpls: argTpls,
	}, &UserConfig{}, nil
}

var _ publisher.User = (*UserConfig)(nil)

type UserConfig struct{}

func (u *UserConfig) SetAuthToken(token string) {}
