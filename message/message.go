package message

type Message struct {
	Text          string
	User          string
	BotUser       string
	ChannelId     string
	ChannelName   string
	ThreadId      string
	InThread      bool
	DirectMessage bool
	Locale        string
}
