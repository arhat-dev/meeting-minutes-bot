package rt

// GeneratorOutput is the output of a generator
type GeneratorOutput struct {
	Messages []*Message
	Data     Optional[string]

	Other []GeneratorOutput
}

type GeneratorInput struct {
	Cmd      string
	Params   string
	Messages []*Message
}
