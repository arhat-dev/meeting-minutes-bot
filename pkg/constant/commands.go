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

	AllBotCommands = append([]string{CommandStart}, VisibleBotCommands...)

	BotCommandShortDescriptions = map[string]string{
		CommandDiscuss: "start a new session around some topic",
		CommandIgnore:  "ignore some message during session",
		CommandInclude: "include extra message to session",
		CommandEnd:     "end current session",

		CommandContinue: "continue previously created session",

		CommandEdit:   "edit session post",
		CommandList:   "list all session posts",
		CommandDelete: "delete certain session post",

		CommandHelp: "show help text",

		CommandStart: "",
	}
)
