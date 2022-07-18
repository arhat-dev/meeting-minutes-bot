package filter

import (
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/rs"
)

const (
	Name = "filter"
)

func init() {
	generator.Register(Name, func() generator.Config { return &Config{} })
}

type Op uint32

const (
	op_Delete Op = iota
	op_Ignore
	op_Edit
)

const (
	opStr_Delete = "delete" // delete this message in chat, implies "ignore"
	opStr_Ignore = "ignore" // ignore this message when generating
	opStr_Edit   = "edit"   // edit this message when generating
)

type OperationSpec struct {
	rs.BaseField

	// MatchText is a regular expression to match full text of a message
	MatchText string `yaml:"matchText"`

	// MatchMediaContentType is a regular expression to match MIME value of a media span
	MatchMediaContentType string `yaml:"matchMediaContentType"`

	OnMatch    string `yaml:"onMatch"`
	OnMismatch string `yaml:"onMismatch"`

	// EditTemplate is a golang template used to edit this message
	EditTemplate string `yaml:"editTemplate"`
}

type Config struct {
	rs.BaseField

	// TODO
	Ops []*OperationSpec `yaml:"ops"`
}

func (c *Config) Create() (generator.Interface, error) {
	return &Driver{}, nil
}
