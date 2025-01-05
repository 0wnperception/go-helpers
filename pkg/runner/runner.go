package runner

import (
	"context"

	"github.com/0wnperception/go-helpers/pkg/log"
)

func Run(appName string, opts ...Option) {
	ctx := context.Background()

	logCtx, logClean := log.NewCtx(ctx, appName, true)
	defer logClean()

	cfg := NewAppConfig(logCtx, opts...)

	log.Info(logCtx, "Starting application "+appName)

	startServices(cfg)

	log.Info(logCtx, "Application "+appName+" is ready!")

	wait(cfg)

	log.Info(logCtx, "Stopping application "+appName)

	onShutdown(cfg)

	gracefulStop(cfg)

	log.Info(logCtx, "Application "+appName+" is stopped!")
}

func startServices(cfg *Config) {
	for i := range cfg.services {
		f := cfg.services[i]
		if f != nil {
			cfg.wg.Add(1)

			go func() {
				f(cfg.ctx)
				cfg.wg.Done()
			}()
		}
	}
}

func wait(cfg *Config) {
	cfg.waitFunc()
}

func onShutdown(cfg *Config) {
	for _, f := range cfg.shutdownHooks {
		if f != nil {
			f(cfg.ctx)
		}
	}
}

func gracefulStop(cfg *Config) {
	cfg.wg.Wait()
	if cfg.cancel != nil {
		cfg.cancel()
	}
}
