package ctxcache

import (
	"context"
	"errors"
	"reflect"
)

var (
	ErrSyncContextNotFound = errors.New("sync context not found in context")
)

// SyncContextKey - ключ для хранения SyncContext в context
type SyncContextKey struct{}

// SyncContext - контекст синхронизации с кэшем результатов
// M - тип метаданных (ExchangeMeta и т.д.)
type SyncContext[M any] struct {
	cache map[reflect.Type]any // кэш синхронизированных данных по типу
	meta  M                    // метаданные
}

// NewSyncContext создает новый SyncContext и добавляет его в контекст
func NewSyncContext[M any](ctx context.Context, meta M) context.Context {
	syncCtx := &SyncContext[M]{
		cache: make(map[reflect.Type]any),
		meta:  meta,
	}
	return setSyncContext(ctx, syncCtx)
}

// getSyncContext получает SyncContext из контекста (внутренняя функция)
func getSyncContext[M any](ctx context.Context) (*SyncContext[M], bool) {
	value := ctx.Value(SyncContextKey{})
	if value == nil {
		return nil, false
	}

	syncCtx, ok := value.(*SyncContext[M])
	if !ok {
		return nil, false
	}

	return syncCtx, true
}

// setSyncContext сохраняет SyncContext в контекст (внутренняя функция)
func setSyncContext[M any](ctx context.Context, syncCtx *SyncContext[M]) context.Context {
	return context.WithValue(ctx, SyncContextKey{}, syncCtx)
}

// GetSyncCached получает данные из кэша по типу
// M - тип метаданных в SyncContext
func GetSyncCached[T, M any](ctx context.Context) (T, bool) {
	var zero T
	syncCtx, ok := getSyncContext[M](ctx)
	if !ok {
		return zero, false
	}

	typ := reflect.TypeOf(zero)
	data, ok := syncCtx.cache[typ]
	if !ok {
		return zero, false
	}

	result, ok := data.(T)
	return result, ok
}

// SetSyncCached сохраняет данные в кэш по типу
// M - тип метаданных в SyncContext
func SetSyncCached[T, M any](ctx context.Context, data T) (context.Context, error) {
	syncCtx, ok := getSyncContext[M](ctx)
	if !ok {
		return nil, ErrSyncContextNotFound
	}

	var zero T
	typ := reflect.TypeOf(zero)
	syncCtx.cache[typ] = data

	return setSyncContext(ctx, syncCtx), nil
}

// GetSyncMeta получает метаданные из SyncContext
func GetSyncMeta[M any](ctx context.Context) (M, bool) {
	var zero M
	syncCtx, ok := getSyncContext[M](ctx)
	if !ok {
		return zero, false
	}
	return syncCtx.meta, true
}
