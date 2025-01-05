package log

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/0wnperception/go-helpers/pkg/types"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Test1(t *testing.T) {
	log := New(true, String("app", "test-app"))
	log.Info("Information message")
	log.Error("Error", Error(errors.New("sample error")),
		Decimal("decimal", types.NewDecimal(12, -1)))
}

func Test2(t *testing.T) {
	log := New(true)
	log.Info("Test One")
	log = log.With(String("field", "value"))
	log.Info("After with")
}

func TestFromContext(t *testing.T) {
	ctx := context.Background()
	logger := New(true)
	logger.Info("Before inject")
	ctx = logger.Inject(ctx)

	FromContext(ctx).Info("Info message")
	ctx = WithFields(ctx, String("F1", "v1"), String("F1", "v2"))
	FromContext(ctx).Info("Info message after")
	ctxChild := WithFields(ctx, Int("childInt", 12))
	FromContext(ctxChild).Info("Log from child context")
	FromContext(ctx).Debug("Log from parent context")
}

func TestNames(t *testing.T) {
	log := New(false)
	log = log.Named("LoggerName")
	log.Info("Info message", String("key", "value1"))
}

func TestEmpty(t *testing.T) {
	log := New(true)
	log = log.Named("LoggerName")
	log.Info("Info message", String("key", "value"), Any("msg", struct{}{}))
}

func TestBuildLoggerFields(t *testing.T) {
	ctx := context.Background()

	l := New(true, String("field1", "val1"), Skip())

	tp := trace.NewTracerProvider()

	closeFunc := func(ctx context.Context) error {
		if e := tp.Shutdown(ctx); e != nil {
			return fmt.Errorf("shutdown trace provider error: %w", e)
		}

		return nil
	}

	defer func() { _ = closeFunc(ctx) }()

	otel.SetTracerProvider(tp)

	// b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)),
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	c, s := tp.Tracer("aaa").Start(ctx, "name")
	defer s.End()

	fields := l.buildLoggerFields(c, String("field1", "val2"),
		Bool("field2", true),
		Int("field2", 2))

	require.Equal(t, 4, len(fields))
	require.Equal(t, "field1", fields[0].Key)
	require.Equal(t, "val2", fields[0].String)
	require.Equal(t, zapcore.Int64Type, fields[1].Type)
	require.Equal(t, "field2", fields[1].Key)
	require.Equal(t, int64(2), fields[1].Integer)
	require.Equal(t, traceID, fields[2].Key)
	require.Equal(t, spanID, fields[3].Key)
}

type Syncer struct {
	err    error
	called bool
}

// SetError sets the error that the Sync method will return.
func (s *Syncer) SetError(err error) {
	s.err = err
}

// Sync records that it was called, then returns the user-supplied error (if
// any).
func (s *Syncer) Sync() error {
	s.called = true
	return s.err
}

// Called reports whether the Sync method was called.
func (s *Syncer) Called() bool {
	return s.called
}

type Buffer struct {
	bytes.Buffer
	Syncer
}

// Lines returns the current buffer contents, split on newlines.
func (b *Buffer) Lines() []string {
	output := strings.Split(b.String(), "\n")
	return output[:len(output)-1]
}

// Stripped returns the current buffer contents with the last trailing newline
// stripped.
func (b *Buffer) Stripped() string {
	return strings.TrimRight(b.String(), "\n")
}

func testTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("123")
}
func TestLog(t *testing.T) {
	var bs Buffer

	encCfg := NewEncoderConfig()
	encCfg.EncodeTime = testTimeEncoder

	enc := zapcore.NewJSONEncoder(encCfg)

	logger := zap.New(zapcore.NewCore(enc, &bs, zap.DebugLevel))

	ll := &Log{
		l:      logger,
		fields: make(map[string]Field),
	}

	ctx := ll.Inject(context.Background())

	require.True(t, DebugEnabled(ctx))
	require.True(t, WarnEnabled(ctx))
	require.True(t, InfoEnabled(ctx))
	require.True(t, ErrorEnabled(ctx))

	tests := []struct {
		f func()
		s string
	}{
		{func() { Info(ctx, "info") }, `{"level":"INFO","@timestamp":"123","message":"info"}`},
		{func() { Debug(ctx, "debug") }, `{"level":"DEBUG","@timestamp":"123","message":"debug"}`},
		{func() { Warn(ctx, "warn") }, `{"level":"WARN","@timestamp":"123","message":"warn"}`},
		{func() { Info(ctx, "info", Int("i", 1)) }, `{"level":"INFO","@timestamp":"123","message":"info","i":1}`},
		{func() { Info(ctx, "info", Int32("i", 1)) }, `{"level":"INFO","@timestamp":"123","message":"info","i":1}`},
		{func() { Info(ctx, "info", Int16("i", 1)) }, `{"level":"INFO","@timestamp":"123","message":"info","i":1}`},
		{func() { Info(ctx, "info", Int64("i", 2)) }, `{"level":"INFO","@timestamp":"123","message":"info","i":2}`},
		{func() { Info(ctx, "info", Int8("i", 3)) }, `{"level":"INFO","@timestamp":"123","message":"info","i":3}`},
		{func() { Info(ctx, "uint", Uint("u", 1)) }, `{"level":"INFO","@timestamp":"123","message":"uint","u":1}`},
		{func() { Info(ctx, "uint", Uint16("u", 1)) }, `{"level":"INFO","@timestamp":"123","message":"uint","u":1}`},
		{func() { Info(ctx, "uint", Uint32("u", 2)) }, `{"level":"INFO","@timestamp":"123","message":"uint","u":2}`},
		{func() { Info(ctx, "uint", Uint64("u", 3)) }, `{"level":"INFO","@timestamp":"123","message":"uint","u":3}`},
		{func() { Info(ctx, "uint", Uint8("u", 4)) }, `{"level":"INFO","@timestamp":"123","message":"uint","u":4}`},
		{func() { Info(ctx, "info", Ints("i", []int{1, 2})) }, `{"level":"INFO","@timestamp":"123","message":"info","i":[1,2]}`},
		{func() { Info(ctx, "info", Int8s("i", []int8{1, 2})) }, `{"level":"INFO","@timestamp":"123","message":"info","i":[1,2]}`},
		{func() { Info(ctx, "info", Int16s("i", []int16{1, 2})) }, `{"level":"INFO","@timestamp":"123","message":"info","i":[1,2]}`},
		{func() { Info(ctx, "info", Int32s("i", []int32{1, 2})) }, `{"level":"INFO","@timestamp":"123","message":"info","i":[1,2]}`},
		{func() { Info(ctx, "info", Int64s("i", []int64{1, 2})) }, `{"level":"INFO","@timestamp":"123","message":"info","i":[1,2]}`},
		{func() { Info(ctx, "uint", Uints("u", []uint{2, 3})) }, `{"level":"INFO","@timestamp":"123","message":"uint","u":[2,3]}`},
		{func() { Info(ctx, "uint", Uint8s("u", []uint8{2, 3})) }, `{"level":"INFO","@timestamp":"123","message":"uint","u":[2,3]}`},
		{func() { Info(ctx, "uint", Uint16s("u", []uint16{2, 3})) }, `{"level":"INFO","@timestamp":"123","message":"uint","u":[2,3]}`},
		{func() { Info(ctx, "uint", Uint32s("u", []uint32{2, 3})) }, `{"level":"INFO","@timestamp":"123","message":"uint","u":[2,3]}`},
		{func() { Info(ctx, "uint", Uint64s("u", []uint64{2, 3})) }, `{"level":"INFO","@timestamp":"123","message":"uint","u":[2,3]}`},
	}

	for _, tt := range tests {
		tt.f()
	}

	lines := bs.Lines()
	for i, tt := range tests {
		require.Equalf(t, tt.s, lines[i], "test index: %d", i)
	}
}
