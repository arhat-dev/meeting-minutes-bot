package bot

import (
	"fmt"
	"strings"

	"arhat.dev/rs"

	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/mbot/pkg/storage"
	"arhat.dev/mbot/pkg/webarchiver"
)

// WorkflowConfig represents a self-contained service for bot
type WorkflowConfig struct {
	rs.BaseField

	CmdMapping rt.CommandsMapping `yaml:"cmdMapping"`

	// Storage config name
	Storage string `yaml:"storage"`

	// WebArchiver config name
	WebArchiver string `yaml:"webarchiver"`

	// Generator config name
	Generator string `yaml:"generator"`

	// Publisher config name
	Publisher string `yaml:"publisher"`
}

func (wfc *WorkflowConfig) Resolve(bctx *BotContext) (ret Workflow, err error) {
	var (
		wa webarchiver.Interface
		st storage.Interface
		gn generator.Interface
	)

	gnConf, ok := bctx.Generators[wfc.Generator]
	if !ok {
		err = fmt.Errorf("unknown generator %q", wfc.Generator)
		return
	}

	gn, err = gnConf.Create()
	if err != nil {
		return
	}

	stConf, ok := bctx.StorageSets[wfc.Storage]
	if !ok {
		err = fmt.Errorf("unknown storage set %q", wfc.Storage)
		return
	}

	st, err = stConf.Create()
	if err != nil {
		return
	}

	if len(wfc.WebArchiver) != 0 {
		waConf, ok := bctx.WebArchivers[wfc.WebArchiver]
		if !ok {
			err = fmt.Errorf("unknown webarchiver %q", wfc.WebArchiver)
			return
		}

		wa, err = waConf.Create()
		if err != nil {
			return
		}
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

		Storage:     st,
		WebArchiver: wa,
		Generator:   gn,

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

	Storage     storage.Interface
	WebArchiver webarchiver.Interface
	Generator   generator.Interface

	pbName        string
	pbFactoryFunc PublisherFactoryFunc
}

func (c *Workflow) PublisherName() string { return c.pbName }
func (c *Workflow) CreatePublisher() (publisher.Interface, publisher.User, error) {
	return c.pbFactoryFunc()
}