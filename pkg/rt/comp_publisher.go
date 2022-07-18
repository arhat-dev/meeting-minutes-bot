package rt

// PublisherOutput is the output of a publisher
type PublisherOutput struct {
	SendMessage Optional[SendMessageOptions]

	Other []PublisherOutput
}
