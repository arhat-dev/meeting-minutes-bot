package constant

// bot commands
const (
	CommandDiscuss = "/discuss"
	CommandIgnore  = "/ignore"
	CommandInclude = "/include"
	CommandEnd     = "/end"

	CommandContinue = "/continue"

	// post management
	CommandEdit   = "/edit"
	CommandList   = "/list"
	CommandDelete = "/delete"

	CommandHelp = "/help"

	CommandStart = "/start"
)

// command helper texts
var (
	VisibleBotCommands = []string{
		CommandDiscuss,
		CommandIgnore,
		CommandInclude,
		CommandEnd,

		CommandContinue,

		CommandEdit,
		CommandList,
		CommandDelete,

		CommandHelp,
	}

	BotCommandShortDescriptions = map[string]string{
		CommandDiscuss: "start a new discussion around some topic",
		CommandIgnore:  "ignore some message during discussion",
		CommandInclude: "include extra message to discussion",
		CommandEnd:     "end current discussion",

		CommandContinue: "continue previously created discussion",

		CommandEdit:   "edit discussion post",
		CommandList:   "list all discussion posts",
		CommandDelete: "delete certain discussion post",

		CommandHelp: "show help text",

		CommandStart: "",
	}
)

func CommandHelpText() string {
	body := ""
	for _, cmd := range VisibleBotCommands {
		body += "<pre>" + cmd + "</pre> - " + BotCommandShortDescriptions[cmd] + "\n"
	}

	return `Usage:

` + body + `
`
}
