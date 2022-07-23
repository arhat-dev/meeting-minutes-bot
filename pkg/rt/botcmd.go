package rt

type BotCmd uint8

const (
	BotCmd_Unknown BotCmd = iota

	BotCmd_Help
	BotCmd_Start

	BotCmd_New
	BotCmd_Resume
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

	case BotCmd_New:
		return BotCmdText_New
	case BotCmd_Resume:
		return BotCmdText_Resume
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

func DefaultBotCommands() BotCommands {
	return BotCommands{
		botCmd_Help:  BotCmdText_Help,
		botCmd_Start: BotCmdText_Start,

		botCmd_New:     BotCmdText_New,
		botCmd_Resume:  BotCmdText_Resume,
		botCmd_Ignore:  BotCmdText_Ignore,
		botCmd_Include: BotCmdText_Include,
		botCmd_End:     BotCmdText_End,
		botCmd_Cancel:  BotCmdText_Cancel,

		botCmd_Edit:   BotCmdText_Edit,
		botCmd_List:   BotCmdText_List,
		botCmd_Delete: BotCmdText_Delete,

		Commands: [botCmd_Count]string{
			BotCmd_Help - 1:  BotCmdText_Help,
			BotCmd_Start - 1: BotCmdText_Start,

			BotCmd_New - 1:     BotCmdText_New,
			BotCmd_Resume - 1:  BotCmdText_Resume,
			BotCmd_Ignore - 1:  BotCmdText_Ignore,
			BotCmd_Include - 1: BotCmdText_Include,
			BotCmd_End - 1:     BotCmdText_End,
			BotCmd_Cancel - 1:  BotCmdText_Cancel,

			BotCmd_Edit - 1:   BotCmdText_Edit,
			BotCmd_List - 1:   BotCmdText_List,
			BotCmd_Delete - 1: BotCmdText_Delete,
		},

		Descriptions: [botCmd_Count]string{
			BotCmd_Help - 1:  BotCmdDesc_Help,
			BotCmd_Start - 1: BotCmdDesc_Start,

			BotCmd_New - 1:     BotCmdDesc_New,
			BotCmd_Resume - 1:  BotCmdDesc_Resume,
			BotCmd_Ignore - 1:  BotCmdDesc_Ignore,
			BotCmd_Include - 1: BotCmdDesc_Include,
			BotCmd_End - 1:     BotCmdDesc_End,
			BotCmd_Cancel - 1:  BotCmdDesc_Cancel,

			BotCmd_Edit - 1:   BotCmdDesc_Edit,
			BotCmd_List - 1:   BotCmdDesc_List,
			BotCmd_Delete - 1: BotCmdDesc_Delete,
		},
	}
}

// BotCommands for runtime bot command handling
type BotCommands struct {
	botCmd_Help  string
	botCmd_Start string

	botCmd_New     string
	botCmd_Resume  string
	botCmd_Ignore  string
	botCmd_Include string
	botCmd_End     string
	botCmd_Cancel  string

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

	case c.botCmd_New:
		return BotCmd_New
	case c.botCmd_Resume:
		return BotCmd_Resume
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

const (
	BotCmdText_Help  = "/help"
	BotCmdText_Start = "/start"

	BotCmdText_New     = "/new"
	BotCmdText_Resume  = "/resume"
	BotCmdText_Ignore  = "/ignore"
	BotCmdText_Include = "/include"
	BotCmdText_End     = "/end"
	BotCmdText_Cancel  = "/cancel"

	BotCmdText_Edit   = "/edit"
	BotCmdText_List   = "/list"
	BotCmdText_Delete = "/delete"
)

const (
	BotCmdDesc_Help  = "show help text"
	BotCmdDesc_Start = ""

	BotCmdDesc_New     = "request a new session around some topic"
	BotCmdDesc_Resume  = "continue previously created session"
	BotCmdDesc_Ignore  = "ignore certain message in this session"
	BotCmdDesc_Include = "include extra message in this session"
	BotCmdDesc_End     = "end current session"
	BotCmdDesc_Cancel  = "cancel current request"

	BotCmdDesc_Edit   = "edit session post"
	BotCmdDesc_List   = "list all session posts"
	BotCmdDesc_Delete = "delete certain session post"
)
