package runner

import (
	"context"
	"sync"
)

type ServiceFunc func(ctx context.Context)
type WaitFunc func()

type Config struct {
	services      []ServiceFunc
	waitFunc      WaitFunc
	shutdownHooks []ServiceFunc

	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func NewAppConfig(ctx context.Context, opts ...Option) *Config {
	ctx, cancel := context.WithCancel(ctx)
	cfg := &Config{
		ctx:    ctx,
		cancel: cancel,
		wg:     &sync.WaitGroup{},
	}

	for _, ops := range opts {
		ops(cfg)
	}

	if cfg.waitFunc == nil {
		cfg.waitFunc = DefaultWaitFunc
	}

	return cfg
}
