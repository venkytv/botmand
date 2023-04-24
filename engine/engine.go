package engine

import (
	"context"
)

type Enginer interface {
	Start(context.Context)
}

type EngineFactoryer interface {
	Config() *Config
	Create(env map[string]string, comm *EngineQueues) Enginer
}

type EngineFactoryLoader interface {
	Load(ctx context.Context, config *Config) EngineFactoryer
}
