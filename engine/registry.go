package engine

import (
	"context"
	"fmt"
	"sync"
)

type EngineRegistry struct {
	engines      map[string]EngineFactoryLoader
	registryLock *sync.RWMutex
}

func NewEngineRegistry() EngineRegistry {
	return EngineRegistry{
		engines:      make(map[string]EngineFactoryLoader),
		registryLock: &sync.RWMutex{},
	}
}

func (er EngineRegistry) Register(name string, loader EngineFactoryLoader) {
	er.registryLock.Lock()
	defer er.registryLock.Unlock()

	er.engines[name] = loader
}

// GetEngineFactory returns a EngineFactoryer for the given engine name
func (er EngineRegistry) GetEngineFactory(ctx context.Context, config *Config) (EngineFactoryer, error) {
	er.registryLock.RLock()
	defer er.registryLock.RUnlock()

	loader, ok := er.engines[config.Engine]
	if !ok {
		return nil, fmt.Errorf("Engine not found: %s", config.Engine)
	}

	return loader.Load(ctx, config), nil
}
