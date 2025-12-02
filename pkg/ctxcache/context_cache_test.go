package ctxcache

import (
	"context"
	"testing"
)

// testMeta - тестовые метаданные
type testMeta struct {
	ID   int
	Name string
}

// testData - тестовые данные для кэша
type testData struct {
	Value int
}

func TestNewSyncContext(t *testing.T) {
	ctx := context.Background()
	meta := testMeta{ID: 1, Name: "test"}

	ctx = NewSyncContext(ctx, meta)

	// Проверяем, что метаданные можно получить
	gotMeta, ok := GetSyncMeta[testMeta](ctx)
	if !ok {
		t.Fatal("GetSyncMeta failed: sync context not found")
	}

	if gotMeta.ID != meta.ID || gotMeta.Name != meta.Name {
		t.Errorf("GetMeta() = %v, want %v", gotMeta, meta)
	}
}

func TestGetSyncCached_NotFound(t *testing.T) {
	ctx := context.Background()
	meta := testMeta{ID: 1, Name: "test"}
	ctx = NewSyncContext(ctx, meta)

	// Пытаемся получить данные, которых нет в кэше
	data, ok := GetSyncCached[[]testData, testMeta](ctx)
	if ok {
		t.Errorf("GetSyncCached() = %v, %v, want false", data, ok)
	}
}

func TestSetSyncCached_GetSyncCached(t *testing.T) {
	ctx := context.Background()
	meta := testMeta{ID: 1, Name: "test"}
	ctx = NewSyncContext(ctx, meta)

	// Сохраняем данные в кэш
	testSlice := []testData{{Value: 1}, {Value: 2}}
	ctx, err := SetSyncCached[[]testData, testMeta](ctx, testSlice)
	if err != nil {
		t.Fatalf("SetSyncCached() error = %v", err)
	}

	// Получаем данные из кэша
	gotData, ok := GetSyncCached[[]testData, testMeta](ctx)
	if !ok {
		t.Fatal("GetSyncCached() failed: data not found in cache")
	}

	if len(gotData) != len(testSlice) {
		t.Errorf("GetSyncCached() len = %d, want %d", len(gotData), len(testSlice))
	}

	for i := range gotData {
		if gotData[i].Value != testSlice[i].Value {
			t.Errorf("GetSyncCached() [%d] = %v, want %v", i, gotData[i], testSlice[i])
		}
	}
}

func TestSetSyncCached_WithoutSyncContext(t *testing.T) {
	ctx := context.Background()

	// Пытаемся сохранить данные без SyncContext
	_, err := SetSyncCached[[]testData, testMeta](ctx, []testData{{Value: 1}})
	if err == nil {
		t.Error("SetSyncCached() error = nil, want ErrSyncContextNotFound")
	}

	if err != ErrSyncContextNotFound {
		t.Errorf("SetSyncCached() error = %v, want ErrSyncContextNotFound", err)
	}
}

func TestGetMeta_WithoutSyncContext(t *testing.T) {
	ctx := context.Background()

	// Пытаемся получить метаданные без SyncContext
	meta, ok := GetSyncMeta[testMeta](ctx)
	if ok {
		t.Errorf("GetSyncMeta() = %v, %v, want false", meta, ok)
	}
}

func TestMultipleDataTypes(t *testing.T) {
	ctx := context.Background()
	meta := testMeta{ID: 1, Name: "test"}
	ctx = NewSyncContext(ctx, meta)

	// Сохраняем разные типы данных
	type type1 struct{ A int }
	type type2 struct{ B string }

	ctx, err := SetSyncCached[type1, testMeta](ctx, type1{A: 10})
	if err != nil {
		t.Fatalf("SetSyncCached(type1) error = %v", err)
	}

	ctx, err = SetSyncCached[type2, testMeta](ctx, type2{B: "hello"})
	if err != nil {
		t.Fatalf("SetSyncCached(type2) error = %v", err)
	}

	// Получаем данные разных типов
	got1, ok := GetSyncCached[type1, testMeta](ctx)
	if !ok {
		t.Fatal("GetSyncCached(type1) failed")
	}
	if got1.A != 10 {
		t.Errorf("GetSyncCached(type1) = %v, want A=10", got1)
	}

	got2, ok := GetSyncCached[type2, testMeta](ctx)
	if !ok {
		t.Fatal("GetSyncCached(type2) failed")
	}
	if got2.B != "hello" {
		t.Errorf("GetSyncCached(type2) = %v, want B=hello", got2)
	}
}

func TestGetMeta(t *testing.T) {
	ctx := context.Background()
	meta := testMeta{ID: 42, Name: "test-meta"}
	ctx = NewSyncContext(ctx, meta)

	gotMeta, ok := GetSyncMeta[testMeta](ctx)
	if !ok {
		t.Fatal("GetSyncMeta() failed")
	}

	if gotMeta.ID != 42 || gotMeta.Name != "test-meta" {
		t.Errorf("GetMeta() = %v, want %v", gotMeta, meta)
	}
}

func TestSetSyncCached_UpdatesContext(t *testing.T) {
	ctx := context.Background()
	meta := testMeta{ID: 1, Name: "test"}
	ctx = NewSyncContext(ctx, meta)

	// Сохраняем данные
	ctx1, err := SetSyncCached[[]testData, testMeta](ctx, []testData{{Value: 1}})
	if err != nil {
		t.Fatalf("SetSyncCached() error = %v", err)
	}

	// Проверяем, что новый контекст содержит данные
	data, ok := GetSyncCached[[]testData, testMeta](ctx1)
	if !ok {
		t.Fatal("GetSyncCached() failed after SetSyncCached")
	}

	if len(data) != 1 || data[0].Value != 1 {
		t.Errorf("GetSyncCached() = %v, want [{Value: 1}]", data)
	}

	// Старый контекст также должен содержать данные, так как SyncContext передается по ссылке
	data, ok = GetSyncCached[[]testData, testMeta](ctx)
	if !ok {
		t.Error("GetSyncCached() on old context should also work (SyncContext is shared by reference)")
	}

	if len(data) != 1 || data[0].Value != 1 {
		t.Errorf("GetSyncCached() on old context = %v, want [{Value: 1}]", data)
	}
}
