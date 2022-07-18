package rt

import "arhat.dev/rs"

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
	ret = DefaultBotCommands()

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
