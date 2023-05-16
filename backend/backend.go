package backend

import "github.com/venkytv/botters/message"

const QBufferSize = 100

type Backender interface {
	Name() string
	Read()
	Post()
	Sanitize(*message.Message) *message.Message
}

type BackendQueues struct {
	MesgQ chan *message.Message
	RespQ chan *message.Message
}

func NewBackendQueues() BackendQueues {
	return BackendQueues{
		MesgQ: make(chan *message.Message, QBufferSize),
		RespQ: make(chan *message.Message, QBufferSize),
	}
}
