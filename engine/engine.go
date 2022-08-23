package engine

import "context"

const QBufferSize = 10

type EngineConfig struct {
	Threaded bool     `yaml:"threaded"`
	Patterns []string `yaml:"patterns"`
}

type Enginer interface {
	Start(context.Context)
}

type EngineFactoryer interface {
	Id() string
	Config() EngineConfig
	Create(env map[string]string, comm *EngineQueues) Enginer
}

type EngineFactoryLoader interface {
	Load(ctx context.Context) []EngineFactoryer
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
