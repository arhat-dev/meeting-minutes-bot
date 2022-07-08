package bot

import (
	"fmt"

	"arhat.dev/rs"

	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/publisher"
	"arhat.dev/meeting-minutes-bot/pkg/storage"
	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"
)

// WorkflowConfig represents a self-contained service for bot
type WorkflowConfig struct {
	rs.BaseField

	CmdMapping CommandsMapping `yaml:"cmdMapping"`

	// StorageSet config name
	StorageSet string `yaml:"storageSet"`

	// WebArchiver config name
	WebArchiver string `yaml:"webarchiver"`

	// Generator config name
	Generator string `yaml:"generator"`

	// Publisher config name
	Publisher string `yaml:"publisher"`
}

func (wfc *WorkflowConfig) Resolve(bctx *Context) (ret Workflow, err error) {
	var (
		wa  webarchiver.Interface
		st  storage.Interface
		gn  generator.Interface
		_pb publisher.Interface
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

	stConf, ok := bctx.StorageSets[wfc.StorageSet]
	if !ok {
		err = fmt.Errorf("unknown storage set %q", wfc.StorageSet)
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
		err = fmt.Errorf("unknown publisher %q", wfc.StorageSet)
		return
	}

	_pb, _, err = pbConf.Create()
	if err != nil {
		err = fmt.Errorf("check publisher %q: %w", wfc.Publisher, err)
		return
	}

	ret = Workflow{
		BotCommands: wfc.CmdMapping.Resovle(),

		Storage:     st,
		WebArchiver: wa,
		Generator:   gn,

		pbName:         _pb.Name(),
		pbRequireLogin: _pb.RequireLogin(),
		pbFactoryFunc:  pbConf.Create,
	}

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
	BotCommands BotCommands

	Storage     storage.Interface
	WebArchiver webarchiver.Interface
	Generator   generator.Interface

	pbName         string
	pbRequireLogin bool
	pbFactoryFunc  PublisherFactoryFunc
}

func (c *Workflow) PublisherName() string       { return c.pbName }
func (c *Workflow) PublisherRequireLogin() bool { return c.pbRequireLogin }
func (c *Workflow) CreatePublisher() (publisher.Interface, publisher.UserConfig, error) {
	return c.pbFactoryFunc()
}
