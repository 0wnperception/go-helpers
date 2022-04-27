package logger

type LoggerConfig struct {
	LogLevel string
	LogPath  string
}

//go:generate minimock -i robot_agent/pkg/logger.LoggerInterface -o ./tests/ -s _mock.go
type LoggerInterface interface {
	Release()
	Debug(msg string, arg string, val string)
	Info(msg string, arg string, val string)
	Warn(msg string, arg string, val string)
	Error(msg string, err error)
	Fatal(msg string, err error)
}
