package bot

import (
	"arhat.dev/rs"
)

const (
	BotCmdText_Help  = "/help"
	BotCmdText_Start = "/start"

	BotCmdText_Discuss  = "/discuss"
	BotCmdText_Continue = "/continue"
	BotCmdText_Ignore   = "/ignore"
	BotCmdText_Include  = "/include"
	BotCmdText_End      = "/end"
	BotCmdText_Cancel   = "/cancel"

	BotCmdText_Edit   = "/edit"
	BotCmdText_List   = "/list"
	BotCmdText_Delete = "/delete"
)

const (
	BotCmdDesc_Help  = "show help text"
	BotCmdDesc_Start = ""

	BotCmdDesc_Discuss  = "request a new session around some topic"
	BotCmdDesc_Continue = "continue previously created session"
	BotCmdDesc_Ignore   = "ignore certain message in this session"
	BotCmdDesc_Include  = "include extra message in this session"
	BotCmdDesc_End      = "end current session"
	BotCmdDesc_Cancel   = "cancel current request"

	BotCmdDesc_Edit   = "edit session post"
	BotCmdDesc_List   = "list all session posts"
	BotCmdDesc_Delete = "delete certain session post"
)

type BotCmd uint8

const (
	BotCmd_Unknown BotCmd = iota

	BotCmd_Help
	BotCmd_Start

	BotCmd_Discuss
	BotCmd_Continue
	BotCmd_Ignore
	BotCmd_Include
	BotCmd_End
	BotCmd_Cancel

	BotCmd_Edit
	BotCmd_List
	BotCmd_Delete

	botCmd_Count = iota - 1
)

func (c BotCmd) String() string {
	switch c {
	case BotCmd_Help:
		return BotCmdText_Help
	case BotCmd_Start:
		return BotCmdText_Start

	case BotCmd_Discuss:
		return BotCmdText_Discuss
	case BotCmd_Continue:
		return BotCmdText_Continue
	case BotCmd_Ignore:
		return BotCmdText_Ignore
	case BotCmd_Include:
		return BotCmdText_Include
	case BotCmd_End:
		return BotCmdText_End
	case BotCmd_Cancel:
		return BotCmdText_Cancel

	case BotCmd_Edit:
		return BotCmdText_Edit
	case BotCmd_List:
		return BotCmdText_List
	case BotCmd_Delete:
		return BotCmdText_Delete
	default:
		return "<unknown>"
	}
}

// BotCommands for runtime bot command handling
type BotCommands struct {
	botCmd_Help  string
	botCmd_Start string

	botCmd_Discuss  string
	botCmd_Continue string
	botCmd_Ignore   string
	botCmd_Include  string
	botCmd_End      string
	botCmd_Cancel   string

	botCmd_Edit   string
	botCmd_List   string
	botCmd_Delete string

	Commands     [botCmd_Count]string
	Descriptions [botCmd_Count]string
}

func (c *BotCommands) TextOf(cmd BotCmd) string {
	if cmd <= 0 || cmd > botCmd_Count {
		return "<unknown>"
	}

	return c.Commands[cmd-1]
}

func (c *BotCommands) DescriptionOf(cmd BotCmd) string {
	if cmd <= 0 || cmd > botCmd_Count {
		return "<unknown>"
	}

	return c.Descriptions[cmd-1]
}

func (c *BotCommands) Parse(x string) BotCmd {
	switch x {
	case c.botCmd_Help:
		return BotCmd_Help
	case c.botCmd_Start:
		return BotCmd_Start

	case c.botCmd_Discuss:
		return BotCmd_Discuss
	case c.botCmd_Continue:
		return BotCmd_Continue
	case c.botCmd_Ignore:
		return BotCmd_Ignore
	case c.botCmd_Include:
		return BotCmd_Include
	case c.botCmd_End:
		return BotCmd_End
	case c.botCmd_Cancel:
		return BotCmd_Cancel

	case c.botCmd_Edit:
		return BotCmd_Edit
	case c.botCmd_List:
		return BotCmd_List
	case c.botCmd_Delete:
		return BotCmd_Delete

	default:
		return BotCmd_Unknown
	}
}

// CommandsMapping maps original meeting-minutes targeted bot commands to workflow specific commands
// for better UX
type CommandsMapping struct {
	rs.BaseField

	Help  *CmdMapping `yaml:"/help"`
	Start *CmdMapping `yaml:"/start"`

	Discuss  *CmdMapping `yaml:"/discuss"`
	Continue *CmdMapping `yaml:"/continue"`
	Ignore   *CmdMapping `yaml:"/ignore"`
	Include  *CmdMapping `yaml:"/include"`
	End      *CmdMapping `yaml:"/end"`
	Cancel   *CmdMapping `yaml:"/cancel"`

	Edit   *CmdMapping `yaml:"/edit"`
	List   *CmdMapping `yaml:"/list"`
	Delete *CmdMapping `yaml:"/delete"`
}

type CmdMapping struct {
	rs.BaseField

	As          string `yaml:"as"`
	Description string `yaml:"description"`
}

func (c CommandsMapping) Resovle() (ret BotCommands) {
	ret = BotCommands{
		botCmd_Help:  BotCmdText_Help,
		botCmd_Start: BotCmdText_Start,

		botCmd_Discuss:  BotCmdText_Discuss,
		botCmd_Continue: BotCmdText_Continue,
		botCmd_Ignore:   BotCmdText_Ignore,
		botCmd_Include:  BotCmdText_Include,
		botCmd_End:      BotCmdText_End,
		botCmd_Cancel:   BotCmdText_Cancel,

		botCmd_Edit:   BotCmdText_Edit,
		botCmd_List:   BotCmdText_List,
		botCmd_Delete: BotCmdText_Delete,

		Commands: [botCmd_Count]string{
			BotCmd_Help - 1:  BotCmdText_Help,
			BotCmd_Start - 1: BotCmdText_Start,

			BotCmd_Discuss - 1:  BotCmdText_Discuss,
			BotCmd_Continue - 1: BotCmdText_Continue,
			BotCmd_Ignore - 1:   BotCmdText_Ignore,
			BotCmd_Include - 1:  BotCmdText_Include,
			BotCmd_End - 1:      BotCmdText_End,
			BotCmd_Cancel - 1:   BotCmdText_Cancel,

			BotCmd_Edit - 1:   BotCmdText_Edit,
			BotCmd_List - 1:   BotCmdText_List,
			BotCmd_Delete - 1: BotCmdText_Delete,
		},

		Descriptions: [botCmd_Count]string{
			BotCmd_Help - 1:  BotCmdDesc_Help,
			BotCmd_Start - 1: BotCmdDesc_Start,

			BotCmd_Discuss - 1:  BotCmdDesc_Discuss,
			BotCmd_Continue - 1: BotCmdDesc_Continue,
			BotCmd_Ignore - 1:   BotCmdDesc_Ignore,
			BotCmd_Include - 1:  BotCmdDesc_Include,
			BotCmd_End - 1:      BotCmdDesc_End,
			BotCmd_Cancel - 1:   BotCmdDesc_Cancel,

			BotCmd_Edit - 1:   BotCmdDesc_Edit,
			BotCmd_List - 1:   BotCmdDesc_List,
			BotCmd_Delete - 1: BotCmdDesc_Delete,
		},
	}

loop:
	for {
		switch {
		case c.Help != nil:
			ret.botCmd_Help = c.Help.As
			ret.Commands[BotCmd_Help-1] = c.Help.As
			ret.Descriptions[BotCmd_Help-1] = c.Help.Description
			c.Help = nil
		case c.Start != nil:
			ret.botCmd_Start = c.Start.As
			ret.Commands[BotCmd_Start-1] = c.Start.As
			ret.Descriptions[BotCmd_Start-1] = c.Start.Description
			c.Start = nil

		case c.Discuss != nil:
			ret.botCmd_Discuss = c.Discuss.As
			ret.Commands[BotCmd_Discuss-1] = c.Discuss.As
			ret.Descriptions[BotCmd_Discuss-1] = c.Discuss.Description
			c.Discuss = nil
		case c.Continue != nil:
			ret.botCmd_Continue = c.Continue.As
			ret.Commands[BotCmd_Continue-1] = c.Continue.As
			ret.Descriptions[BotCmd_Continue-1] = c.Continue.Description
			c.Continue = nil
		case c.Ignore != nil:
			ret.botCmd_Ignore = c.Ignore.As
			ret.Commands[BotCmd_Ignore-1] = c.Ignore.As
			ret.Descriptions[BotCmd_Ignore-1] = c.Ignore.Description
			c.Ignore = nil
		case c.Include != nil:
			ret.botCmd_Include = c.Include.As
			ret.Commands[BotCmd_Include-1] = c.Include.As
			ret.Descriptions[BotCmd_Include-1] = c.Include.Description
			c.Include = nil
		case c.End != nil:
			ret.botCmd_End = c.End.As
			ret.Commands[BotCmd_End-1] = c.End.As
			ret.Descriptions[BotCmd_End-1] = c.End.Description
			c.End = nil
		case c.Cancel != nil:
			ret.botCmd_Cancel = c.Cancel.As
			ret.Commands[BotCmd_Cancel-1] = c.Cancel.As
			ret.Descriptions[BotCmd_Cancel-1] = c.Cancel.Description
			c.Cancel = nil

		case c.Edit != nil:
			ret.botCmd_Edit = c.Edit.As
			ret.Commands[BotCmd_Edit-1] = c.Edit.As
			ret.Descriptions[BotCmd_Edit-1] = c.Edit.Description
			c.Edit = nil
		case c.List != nil:
			ret.botCmd_List = c.List.As
			ret.Commands[BotCmd_List-1] = c.List.As
			ret.Descriptions[BotCmd_List-1] = c.List.Description
			c.List = nil
		case c.Delete != nil:
			ret.botCmd_Delete = c.Delete.As
			ret.Commands[BotCmd_Delete-1] = c.Delete.As
			ret.Descriptions[BotCmd_Delete-1] = c.Delete.Description
			c.Delete = nil

		default:
			break loop
		}
	}

	return
}
