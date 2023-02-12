package runner

type Option func(cfg *Config)

func WithServices(fs ...ServiceFunc) Option {
	return func(cfg *Config) {
		cfg.services = append(cfg.services, fs...)
	}
}

func WithShutdownHooks(fs ...ServiceFunc) Option {
	return func(cfg *Config) {
		cfg.shutdownHooks = append(cfg.shutdownHooks, fs...)
	}
}

func WithWaitFunc(f WaitFunc) Option {
	return func(cfg *Config) {
		cfg.waitFunc = f
	}
}
