package bot

import (
	"fmt"
	"strings"

	"arhat.dev/rs"

	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/mbot/pkg/storage"
)

// WorkflowConfig represents a self-contained service for bot
type WorkflowConfig struct {
	rs.BaseField

	CmdMapping rt.CommandsMapping `yaml:"cmdMapping"`

	// Storage config name
	Storage string `yaml:"storage"`

	// Generator config name
	Generator string `yaml:"generator"`

	// Publisher config name
	Publisher string `yaml:"publisher"`
}

func (wfc *WorkflowConfig) Resolve(bctx *CreationContext) (ret Workflow, err error) {
	var (
		st storage.Interface
		gn generator.Interface

		ok bool
	)

	gn, ok = bctx.Generators[wfc.Generator]
	if !ok {
		err = fmt.Errorf("unknown generator %q", wfc.Generator)
		return
	}

	st, ok = bctx.Storage[wfc.Storage]
	if !ok {
		err = fmt.Errorf("unknown storage %q", wfc.Storage)
		return
	}

	pbConf, ok := bctx.Publishers[wfc.Publisher]
	if !ok {
		err = fmt.Errorf("unknown publisher %q", wfc.Storage)
		return
	}

	_, _, err = pbConf.Create()
	if err != nil {
		err = fmt.Errorf("check publisher creation %q: %w", wfc.Publisher, err)
		return
	}

	ret = Workflow{
		BotCommands: wfc.CmdMapping.Resovle(),

		Storage:   st,
		Generator: gn,

		pbFactoryFunc: pbConf.Create,
	}

	ret.pbName, _, _ = strings.Cut(wfc.Publisher, ":")

	return
}

// WorkflowSet is a collection of all workflows for a single bot
type WorkflowSet struct {
	index     map[string]int
	Workflows []Workflow
}

func (w *WorkflowSet) WorkflowFor(cmd string) (ret *Workflow, ok bool) {
	idx, ok := w.index[cmd]
	if !ok {
		return
	}

	return &w.Workflows[idx], true
}

// Workflow contains all runtime components for a workflow
type Workflow struct {
	BotCommands rt.BotCommands

	Storage   storage.Interface
	Generator generator.Interface

	pbName        string
	pbFactoryFunc PublisherFactoryFunc
}

func (c *Workflow) PublisherName() string { return c.pbName }
func (c *Workflow) CreatePublisher() (publisher.Interface, publisher.User, error) {
	return c.pbFactoryFunc()
}
