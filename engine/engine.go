package engine

import "context"

const QBufferSize = 10

type Enginer interface {
	Name() string
	Start(context.Context)
}

type EngineQueues struct {
	ReadQ  chan string
	WriteQ chan string
}

func NewEngineQueues() EngineQueues {
	return EngineQueues{
		ReadQ:  make(chan string, QBufferSize),
		WriteQ: make(chan string, QBufferSize),
	}
}
