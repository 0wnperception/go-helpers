package zap_wrapper

import (
	"local_agent/pkg/logger"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	Core *zap.Logger
	file *os.File
}

func NewZapLogger(cfg *logger.LoggerConfig) (*ZapLogger, error) {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "module",
		TimeKey:        "time",
		CallerKey:      "call",
		EncodeCaller:   zapcore.FullCallerEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	var (
		encoder     zapcore.Encoder
		writeSincer zapcore.WriteSyncer
		level       zapcore.Level
		file        *os.File
		err         error
	)
	switch cfg.LogLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}
	if cfg.LogPath == "" {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
		writeSincer = os.Stdout
	} else {
		file, err = os.Create(cfg.LogPath)
		if err != nil {
			return nil, err
		}
		encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
		encoder = zapcore.NewJSONEncoder(encoderCfg)
		writeSincer = zapcore.AddSync(file)
	}
	return &ZapLogger{
		Core: zap.New(zapcore.NewCore(encoder, writeSincer, level), zap.AddCaller()),
		file: file,
	}, nil
}

func (l *ZapLogger) Release() {
	l.Core.Sync()
	l.file.Close()
}

func (l *ZapLogger) Debug(msg string, arg string, val string) {
	if arg != "" {
		l.Core.Debug(msg, zap.String(arg, val))
	} else {
		l.Core.Debug(msg)
	}
}

func (l *ZapLogger) Info(msg string, arg string, val string) {
	if arg != "" {
		l.Core.Info(msg, zap.String(arg, val))
	} else {
		l.Core.Info(msg)
	}
}

func (l *ZapLogger) Warn(msg string, arg string, val string) {
	if arg != "" {
		l.Core.Warn(msg, zap.String(arg, val))
	} else {
		l.Core.Warn(msg)
	}
}

func (l *ZapLogger) Error(msg string, err error) {
	l.Core.Error(msg, zap.Error(err))
}

func (l *ZapLogger) Fatal(msg string, err error) {
	l.Core.Fatal(msg, zap.Error(err))
}
