// //nolint: dupl,gofumpt,gochecknoglobals,nestif,errcheck,revive,goconst,mnd
package log

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/0wnperception/go-helpers/pkg/types"
)

const (
	samplingInitial    = 100
	samplingThereafter = 100
	nullString         = "<null>"
	traceID            = "traceId"
	spanID             = "spanId"

	EnvKey     = "env"
	SystemKey  = "system"
	InstKey    = "inst"
	VersionKey = "@version"

	SageVarEnv    = "SAGE_ENV"
	SageVarSystem = "SAGE_SYSTEM"
	SageVarInst   = "SAGE_INST"

	indexCapacity = 16
)

type ctxKey struct{}

var indexesPool = sync.Pool{
	New: func() any {
		return make([]fieldIndex, 0, indexCapacity)
	},
}

var (
	nop       = &Log{l: zap.NewNop()}
	loggerKey = ctxKey{}
)

func Nop() *Log {
	return nop
}

type Log struct {
	l      *zap.Logger
	fields map[string]Field
}

type Field struct {
	field zap.Field
}

func (f Field) GetKey() string {
	return f.field.Key
}

func WithFields(ctx context.Context, fields ...Field) context.Context {
	if ln := len(fields); ln > 0 {
		if l, ok := ctx.Value(loggerKey).(*Log); ok {
			l = l.With(fields...)

			return l.Inject(ctx)
		}
	}

	return ctx
}

func (l *Log) With(fields ...Field) *Log {
	fieldsLen := len(fields)
	if fieldsLen <= 0 {
		return l
	}

	newLog := &Log{
		l:      l.l,
		fields: make(map[string]Field, len(fields)),
	}

	for k, v := range l.fields {
		newLog.fields[k] = v
	}

	for i := range fields {
		newLog.fields[fields[i].GetKey()] = fields[i]
	}

	return newLog
}

func (l *Log) Named(s string) *Log {
	if s == "" {
		return l
	}

	newLog := &Log{
		l:      l.l.Named(s),
		fields: make(map[string]Field),
	}

	if l.fields != nil {
		for k, v := range l.fields {
			newLog.fields[k] = v
		}
	}

	return newLog
}

func String(key, val string) Field {
	return Field{field: zap.String(key, val)}
}

func Stringp(key string, val *string) Field {
	return Field{field: zap.Stringp(key, val)}
}

func Bool(key string, val bool) Field {
	return Field{field: zap.Bool(key, val)}
}

func Boolp(key string, val *bool) Field {
	return Field{field: zap.Boolp(key, val)}
}

func ByteString(key string, val []byte) Field {
	return Field{field: zap.ByteString(key, val)}
}

func Float64(key string, val float64) Field {
	return Field{field: zap.Float64(key, val)}
}

func Float64p(key string, val *float64) Field {
	return Field{field: zap.Float64p(key, val)}
}

func Float32(key string, val float32) Field {
	return Field{field: zap.Float32(key, val)}
}

func Float32p(key string, val *float32) Field {
	return Field{field: zap.Float32p(key, val)}
}

func Int(key string, val int) Field {
	return Field{field: zap.Int(key, val)}
}

func Intp(key string, val *int) Field {
	return Field{field: zap.Intp(key, val)}
}

func Int32(key string, val int32) Field {
	return Field{field: zap.Int32(key, val)}
}

func Int32p(key string, val *int32) Field {
	return Field{field: zap.Int32p(key, val)}
}

func Int64(key string, val int64) Field {
	return Field{field: zap.Int64(key, val)}
}

func Int16(key string, val int16) Field {
	return Field{field: zap.Int16(key, val)}
}

func Int16p(key string, val *int16) Field {
	return Field{field: zap.Int16p(key, val)}
}

func Int8(key string, val int8) Field {
	return Field{field: zap.Int8(key, val)}
}

func Int8p(key string, val *int8) Field {
	return Field{field: zap.Int8p(key, val)}
}

func Uint(key string, val uint) Field {
	return Field{field: zap.Uint(key, val)}
}

func Uintp(key string, val *uint) Field {
	return Field{field: zap.Uintp(key, val)}
}

func Uint8(key string, val uint8) Field {
	return Field{field: zap.Uint8(key, val)}
}

func Uint8p(key string, val *uint8) Field {
	return Field{field: zap.Uint8p(key, val)}
}

func Object(key string, m zapcore.ObjectMarshaler) Field {
	return Field{field: zap.Object(key, m)}
}

func Uint16(key string, val uint16) Field {
	return Field{field: zap.Uint16(key, val)}
}

func Uint32(key string, val uint32) Field {
	return Field{field: zap.Uint32(key, val)}
}

func Uint32p(key string, val *uint32) Field {
	return Field{field: zap.Uint32p(key, val)}
}

func Uint64(key string, val uint64) Field {
	return Field{field: zap.Uint64(key, val)}
}

func Uint64p(key string, val *uint64) Field {
	return Field{field: zap.Uint64p(key, val)}
}

func Stringer(key string, val fmt.Stringer) Field {
	return Field{field: zap.Stringer(key, val)}
}

func Time(key string, val time.Time) Field {
	return Field{field: zap.Time(key, val)}
}

func Timep(key string, val *time.Time) Field {
	return Field{field: zap.Timep(key, val)}
}

func Binary(key string, val []byte) Field {
	return Field{field: zap.Binary(key, val)}
}

func Duration(key string, val time.Duration) Field {
	return Field{field: zap.Duration(key, val)}
}

func Durationp(key string, val *time.Duration) Field {
	return Field{field: zap.Durationp(key, val)}
}

func Skip() Field {
	return Field{field: zap.Skip()}
}

func Decimal(key string, d types.Decimal) Field {
	return Field{field: zap.Stringer(key, d)}
}

func OptDecimal(key string, d types.OptDecimal) Field {
	if d.Defined {
		return Field{field: zap.Stringer(key, d.V)}
	}

	return Field{field: zap.String(key, nullString)}
}

func OptTime(key string, d types.OptTime) Field {
	if d.Defined {
		return Field{field: zap.Time(key, d.V)}
	}

	return Field{field: zap.String(key, nullString)}
}

func NamedError(key string, err error) Field {
	return Field{field: zap.NamedError(key, err)}
}

func Error(err error) Field {
	return NamedError("error", err)
}

func Strings(key string, ss []string) Field {
	return Field{field: zap.Strings(key, ss)}
}

func Errors(key string, errs []error) Field {
	return Field{field: zap.Errors(key, errs)}
}

func Bools(key string, bb []bool) Field {
	return Field{field: zap.Bools(key, bb)}
}

func Ints(key string, ints []int) Field {
	return Field{field: zap.Ints(key, ints)}
}

func Int32s(key string, ints []int32) Field {
	return Field{field: zap.Int32s(key, ints)}
}

func Int64s(key string, ints []int64) Field {
	return Field{field: zap.Int64s(key, ints)}
}

func Int16s(key string, ints []int16) Field {
	return Field{field: zap.Int16s(key, ints)}
}

func Int8s(key string, ints []int8) Field {
	return Field{field: zap.Int8s(key, ints)}
}

func Uints(key string, ints []uint) Field {
	return Field{field: zap.Uints(key, ints)}
}

func Uint32s(key string, ints []uint32) Field {
	return Field{field: zap.Uint32s(key, ints)}
}

func Uint64s(key string, ints []uint64) Field {
	return Field{field: zap.Uint64s(key, ints)}
}

func Uint16s(key string, ints []uint16) Field {
	return Field{field: zap.Uint16s(key, ints)}
}

func Uint8s(key string, ints []uint8) Field {
	return Field{field: zap.Uint8s(key, ints)}
}

func Any(key string, val any) Field {
	return Field{field: zap.Any(key, val)}
}

func Reflect(key string, val any) Field {
	return Field{field: zap.Reflect(key, val)}
}

func Stack(key string) Field {
	return Field{field: zap.Stack(key)}
}

func Times(key string, ts []time.Time) Field {
	return Field{field: zap.Times(key, ts)}
}

type fieldIndex struct {
	name  string
	index int
}

func indexOf(indexes []fieldIndex, name string) int {
	for _, f := range indexes {
		if f.name == name {
			return f.index
		}
	}

	return -1
}

func appendField(f zap.Field, result []zap.Field, indexes []fieldIndex) ([]zap.Field, []fieldIndex) {
	if f.Type == zapcore.SkipType {
		return result, indexes
	}

	index := indexOf(indexes, f.Key)
	if index >= 0 {
		result[index] = f
	} else {
		index = len(result)
		result = append(result, f)
		indexes = append(indexes, fieldIndex{name: f.Key, index: index})
	}

	return result, indexes
}

// buildLoggerFields - дедублицирует поля записи лога.
func (l *Log) buildLoggerFields(ctx context.Context, fields ...Field) []zap.Field {
	result := make([]zap.Field, 0, len(fields)+len(l.fields)+2)

	indexes, ok := indexesPool.Get().([]fieldIndex)
	if !ok {
		panic("invalid struct in pool")
	}

	for _, f := range l.fields {
		result, indexes = appendField(f.field, result, indexes)
	}

	for _, f := range fields {
		result, indexes = appendField(f.field, result, indexes)
	}

	if spCtx := trace.SpanFromContext(ctx).SpanContext(); spCtx.IsValid() {
		result, indexes = appendField(zap.String(traceID, spCtx.TraceID().String()), result, indexes)
		result, indexes = appendField(zap.String(spanID, spCtx.SpanID().String()), result, indexes)
	}

	if cap(indexes) == indexCapacity {
		clear(indexes)
		//nolint:staticcheck
		indexesPool.Put(indexes[:0])
	}

	return result
}

func (l *Log) Error(msg string, fields ...Field) {
	if l.l.Core().Enabled(zapcore.ErrorLevel) {
		l.l.Error(msg, l.buildLoggerFields(context.Background(), fields...)...)
	}
}

func (l *Log) Debug(msg string, fields ...Field) {
	if l.l.Core().Enabled(zapcore.DebugLevel) {
		l.l.Debug(msg, l.buildLoggerFields(context.Background(), fields...)...)
	}
}

func (l *Log) Info(msg string, fields ...Field) {
	if l.l.Core().Enabled(zapcore.InfoLevel) {
		l.l.Info(msg, l.buildLoggerFields(context.Background(), fields...)...)
	}
}

func (l *Log) Warn(msg string, fields ...Field) {
	if l.l.Core().Enabled(zapcore.WarnLevel) {
		l.l.Warn(msg, l.buildLoggerFields(context.Background(), fields...)...)
	}
}

func (l *Log) Fatal(msg string, fields ...Field) {
	l.l.Fatal(msg, l.buildLoggerFields(context.Background(), fields...)...)
}

func (l *Log) Close() {
	_ = l.l.Sync()
}

func (l *Log) Z() *zap.Logger {
	return l.l
}

func (l *Log) DebugEnabled() bool {
	return l.l.Core().Enabled(zapcore.DebugLevel)
}

func DebugEnabled(ctx context.Context) bool {
	if v := ctx.Value(loggerKey); v != nil {
		if l, ok := v.(*Log); ok {
			return l.DebugEnabled()
		}
	}

	return false
}

func (l *Log) WarnEnabled() bool {
	return l.l.Core().Enabled(zapcore.WarnLevel)
}

func WarnEnabled(ctx context.Context) bool {
	if v := ctx.Value(loggerKey); v != nil {
		if l, ok := v.(*Log); ok {
			return l.WarnEnabled()
		}
	}

	return false
}

func (l *Log) InfoEnabled() bool {
	return l.l.Core().Enabled(zapcore.InfoLevel)
}

func InfoEnabled(ctx context.Context) bool {
	if v := ctx.Value(loggerKey); v != nil {
		if l, ok := v.(*Log); ok {
			return l.InfoEnabled()
		}
	}

	return false
}

func (l *Log) ErrorEnabled() bool {
	return l.l.Core().Enabled(zapcore.ErrorLevel)
}

func ErrorEnabled(ctx context.Context) bool {
	if v := ctx.Value(loggerKey); v != nil {
		if l, ok := v.(*Log); ok {
			return l.ErrorEnabled()
		}
	}

	return false
}

// NewCtx creates new logger and put it into context. Returns new context and close logger function.
// //nolint: nonamedreturns
func NewCtx(ctx context.Context, name string, debug bool, fields ...Field) (newCtx context.Context, cleanFn func()) {
	l := New(debug, fields...)
	l.l = l.l.Named(name)
	c := l.Inject(ctx)

	return c, l.Close
}

func WithCallerSkip(ctx context.Context, skip int) context.Context {
	if l, ok := ctx.Value(loggerKey).(*Log); ok {
		return l.WithCallerSkip(skip).Inject(ctx)
	}

	return ctx
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

	ll := &Log{
		l:      logger,
		fields: make(map[string]Field),
	}

	for i := range fields {
		ll.fields[fields[i].GetKey()] = fields[i]
	}

	return ll
}

func (l *Log) WithCallerSkip(skip int) *Log {
	newLog := &Log{
		l:      l.l.WithOptions(zap.AddCallerSkip(skip)),
		fields: make(map[string]Field),
	}

	if l.fields != nil {
		for k, v := range l.fields {
			newLog.fields[k] = v
		}
	}

	return newLog
}

// Inject logger to context and returns new context.
func (l *Log) Inject(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext returns logger from context.
func FromContext(ctx context.Context) *Log {
	if l, ok := ctx.Value(loggerKey).(*Log); ok {
		span := trace.SpanFromContext(ctx)

		if spCtx := span.SpanContext(); spCtx.IsValid() {
			return l.With(String(traceID, spCtx.TraceID().String()), String(spanID, spCtx.SpanID().String()))
		}

		return l
	}

	return Nop()
}

func Err(ctx context.Context, msg string, fields ...Field) {
	if l, ok := ctx.Value(loggerKey).(*Log); ok {
		if l.l.Core().Enabled(zapcore.ErrorLevel) {
			l.l.Error(msg, l.buildLoggerFields(ctx, fields...)...)
		}
	}
}

func Debug(ctx context.Context, msg string, fields ...Field) {
	if l, ok := ctx.Value(loggerKey).(*Log); ok {
		if l.l.Core().Enabled(zapcore.DebugLevel) {
			l.l.Debug(msg, l.buildLoggerFields(ctx, fields...)...)
		}
	}
}

func Info(ctx context.Context, msg string, fields ...Field) {
	if l, ok := ctx.Value(loggerKey).(*Log); ok {
		if l.l.Core().Enabled(zapcore.InfoLevel) {
			l.l.Info(msg, l.buildLoggerFields(ctx, fields...)...)
		}
	}
}

func Warn(ctx context.Context, msg string, fields ...Field) {
	if l, ok := ctx.Value(loggerKey).(*Log); ok {
		if l.l.Core().Enabled(zapcore.WarnLevel) {
			l.l.Warn(msg, l.buildLoggerFields(ctx, fields...)...)
		}
	}
}

func Fatal(ctx context.Context, msg string, fields ...Field) {
	if l, ok := ctx.Value(loggerKey).(*Log); ok {
		l.l.Fatal(msg, l.buildLoggerFields(ctx, fields...)...)
	} else {
		panic(msg)
	}
}

func Panic(ctx context.Context, msg string, fields ...Field) {
	if l, ok := ctx.Value(loggerKey).(*Log); ok {
		if l.l.Core().Enabled(zapcore.PanicLevel) {
			l.l.Panic(msg, l.buildLoggerFields(ctx, fields...)...)
		}
	}
}

func UTCTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.UTC().Format("2006-01-02T15:04:05.000Z0700"))
}

func NewEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "@timestamp",
		LevelKey:       "level",
		NameKey:        "application",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     UTCTimeEncoder,
		EncodeDuration: MillisecondDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func NewConfig() zap.Config {
	return zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    NewEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
	}
}

func MillisecondDurationEncoder(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendFloat64(float64(d) / float64(time.Millisecond))
}
