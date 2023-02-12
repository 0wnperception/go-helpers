package log

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	traceID = "traceId"
)

type ctxKey struct{}

var (
	nop       = &Log{l: zap.NewNop()}
	loggerKey = ctxKey{}
)

func Nop() *Log {
	return nop
}

// FromContext returns logger from context.
func FromContext(ctx context.Context) *Log {
	if l, ok := ctx.Value(loggerKey).(*Log); ok {
		span := trace.SpanFromContext(ctx)
		if !span.IsRecording() {
			return l
		}

		if spCtx := span.SpanContext(); spCtx.IsValid() {
			return l.With(String(traceID, spCtx.TraceID().String()))
		}

		return l
	}

	return Nop()
}

type Log struct {
	l      *zap.Logger
	fields []Field
}

func (l *Log) Named(s string) *Log {
	if s == "" {
		return l
	}

	return &Log{l: l.l.Named(s), fields: append(l.fields[:0:0], l.fields...)}
}

func (l *Log) Close() {
	_ = l.l.Sync()
}

func (l *Log) Inject(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func (l *Log) With(fields ...Field) *Log {
	fieldsLen := len(fields)
	if fieldsLen <= 0 {
		return l
	}

	newLog := &Log{}

	newLog.l = l.l
	newLog.fields = make([]Field, 0, fieldsLen+len(l.fields))
	newLog.fields = append(newLog.fields, l.fields...)
	newLog.fields = append(newLog.fields, fields...)

	return newLog
}

func (l *Log) buildLoggerFields(fields ...Field) []zap.Field {
	ln := len(fields) + len(l.fields)
	logFields := make([]zap.Field, 0, ln)

	for i := range l.fields {
		logFields = append(logFields, l.fields[i].field)
	}

	for i := range fields {
		logFields = append(logFields, fields[i].field)
	}

	return logFields
}

func NewCtx(ctx context.Context, name string, debug bool, fields ...Field) (newCtx context.Context, closeFn func()) {
	l := New(debug, fields...).Named(name)
	c := l.Inject(ctx)

	return c, l.Close
}

func New(debug bool, fields ...Field) *Log {
	cfg := NewConfig()

	if debug {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		cfg.Development = true
	}

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	logger = logger.WithOptions(zap.AddCallerSkip(1))

	if ln := len(fields); ln > 0 {
		zapFields := make([]zap.Field, 0, ln)

		for i := range fields {
			zapFields = append(zapFields, fields[i].field)
		}

		logger = logger.With(zapFields...)
	}

	return &Log{l: logger}
}

func addTraceId(ctx context.Context, fields ...Field) []Field {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return fields
	}

	if spCtx := span.SpanContext(); spCtx.IsValid() {
		return append(fields, String(traceID, spCtx.TraceID().String()))
	}

	return fields
}

func WithFields(ctx context.Context, fields ...Field) context.Context {
	if ln := len(fields); ln > 0 {
		l := FromContext(ctx)
		l = l.With(fields...)

		return l.Inject(ctx)
	}

	return ctx
}

func Err(ctx context.Context, msg string, fields ...Field) {
	if l, ok := ctx.Value(loggerKey).(*Log); ok {
		if l.l.Core().Enabled(zapcore.ErrorLevel) {
			l.l.Error(msg, l.buildLoggerFields(addTraceId(ctx, fields...)...)...)
		}
	}
}

func Debug(ctx context.Context, msg string, fields ...Field) {
	if l, ok := ctx.Value(loggerKey).(*Log); ok {
		if l.l.Core().Enabled(zapcore.DebugLevel) {
			l.l.Debug(msg, l.buildLoggerFields(addTraceId(ctx, fields...)...)...)
		}
	}
}

func Info(ctx context.Context, msg string, fields ...Field) {
	if l, ok := ctx.Value(loggerKey).(*Log); ok {
		if l.l.Core().Enabled(zapcore.InfoLevel) {
			l.l.Info(msg, l.buildLoggerFields(addTraceId(ctx, fields...)...)...)
		}
	}
}

func Warn(ctx context.Context, msg string, fields ...Field) {
	if l, ok := ctx.Value(loggerKey).(*Log); ok {
		if l.l.Core().Enabled(zapcore.WarnLevel) {
			l.l.Warn(msg, l.buildLoggerFields(addTraceId(ctx, fields...)...)...)
		}
	}
}

func Fatal(ctx context.Context, msg string, fields ...Field) {
	if l, ok := ctx.Value(loggerKey).(*Log); ok {
		l.l.Fatal(msg, l.buildLoggerFields(addTraceId(ctx, fields...)...)...)
	} else {
		panic(msg)
	}
}
