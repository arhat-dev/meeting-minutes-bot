package constant

// bot commands
const (
	CommandDiscuss = "/discuss"
	CommandIgnore  = "/ignore"
	CommandInclude = "/include"
	CommandEnd     = "/end"
	CommandCancel  = "/cancel"

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
		CommandCancel,

		CommandContinue,

		CommandEdit,
		CommandList,
		CommandDelete,

		CommandHelp,
	}

	AllBotCommands = append([]string{CommandStart}, VisibleBotCommands...)

	BotCommandShortDescriptions = map[string]string{
		CommandDiscuss: "request a new session around some topic",
		CommandIgnore:  "ignore some message during session",
		CommandInclude: "include extra message to session",
		CommandEnd:     "end current session",
		CommandCancel:  "cancel current request",

		CommandContinue: "continue previously created session",

		CommandEdit:   "request edit session post",
		CommandList:   "request list all session posts",
		CommandDelete: "request delete certain session post",

		CommandHelp: "show help text",

		CommandStart: "",
	}
)
