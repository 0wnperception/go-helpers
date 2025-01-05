package zap_log

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogLevelType zapcore.Level

func (l LogLevelType) Original() zapcore.Level {
	return zapcore.Level(l)
}

const (
	LogLevelTypeDebug LogLevelType = LogLevelType(zapcore.DebugLevel)
	LogLevelTypeInfo  LogLevelType = LogLevelType(zapcore.InfoLevel)
	LogLevelTypeWarn  LogLevelType = LogLevelType(zapcore.WarnLevel)
	LogLevelTypeError LogLevelType = LogLevelType(zapcore.ErrorLevel)
)

type ZapLoggerConfig struct {
	LogLevel LogLevelType
	LogPath  string
}

type ZapLogger struct {
	core *zap.Logger
	file *os.File
}

func NewZapLogger(cfg ZapLoggerConfig) (*ZapLogger, error) {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey: "msg",
		LevelKey:   "level",
		NameKey:    "module",
		TimeKey:    "time",
		// CallerKey:      "call",
		// EncodeCaller:   zapcore.FullCallerEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	var (
		encoder     zapcore.Encoder
		writeSincer zapcore.WriteSyncer
		file        *os.File
		err         error
	)

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
		core: zap.New(zapcore.NewCore(encoder, writeSincer, cfg.LogLevel.Original()), zap.AddCaller()),
		file: file,
	}, nil
}

func (l *ZapLogger) Utilize() {
	l.core.Sync()
	if l.file != nil {
		l.file.Close()
	}
}

func (l *ZapLogger) Debug(v ...any) {
	l.core.Debug(fmt.Sprint(v...))
}

func (l *ZapLogger) Debugf(format string, v ...any) {
	l.core.Debug(fmt.Sprintf(format, v...))
}

func (l *ZapLogger) Info(v ...any) {
	l.core.Info(fmt.Sprint(v...))
}

func (l *ZapLogger) Infof(format string, v ...any) {
	l.core.Info(fmt.Sprintf(format, v...))
}

func (l *ZapLogger) Warn(v ...any) {
	l.core.Warn(fmt.Sprint(v...))
}

func (l *ZapLogger) Warnf(format string, v ...any) {
	l.core.Warn(fmt.Sprintf(format, v...))
}

func (l *ZapLogger) Error(v ...any) {
	l.core.Error(fmt.Sprint(v...))
}

func (l *ZapLogger) Errorf(format string, v ...any) {
	l.core.Error(fmt.Sprintf(format, v...))
}

func (l *ZapLogger) Fatal(v ...any) {
	l.core.Fatal(fmt.Sprint(v...))
}

func (l *ZapLogger) Fatalf(format string, v ...any) {
	l.core.Fatal(fmt.Sprintf(format, v...))
}
