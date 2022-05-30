package logger

type LoggerConfig struct {
	LogLevel string
	LogPath  string
}

//go:generate minimock -i robot_agent/pkg/logger.Logger -o ./tests/ -s _mock.go
type Logger interface {
	Debug(v ...any)
	Debugf(format string, v ...any)
	Info(v ...any)
	Infof(format string, v ...any)
	Warn(v ...any)
	Warnf(format string, v ...any)
	Error(v ...any)
	Errorf(format string, v ...any)
	Fatal(v ...any)
	Fatalf(format string, v ...any)
}
