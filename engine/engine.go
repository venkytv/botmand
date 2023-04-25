package engine

import (
	"context"
	"io"
)

type Enginer interface {
	Setup(context.Context) (io.WriteCloser, io.ReadCloser, io.ReadCloser, error)
	Start(context.Context) error
	Wait(context.Context) error
}

type EngineFactoryer interface {
	Config() *Config
	Create(env map[string]string) Enginer
}

type EngineFactoryLoader interface {
	Load(ctx context.Context, config *Config) EngineFactoryer
}
