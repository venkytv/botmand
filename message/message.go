package message

type Message struct {
	Text        string
	User        string
	ChannelId   string
	ChannelName string
	ThreadId    string
	InThread    bool
	Locale      string
}
