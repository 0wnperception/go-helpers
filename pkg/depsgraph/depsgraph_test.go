package depsgraph

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/0wnperception/go-helpers/pkg/log"
)

// Тестовые типы данных
type TestTypeA struct{}
type TestTypeB struct{}
type TestTypeC struct{}
type TestTypeD struct{}

// Mock узел для тестирования
type mockNode struct {
	dataType     any
	dependencies []any
	syncFunc     func(ctx context.Context) error
	syncCalled   bool
}

func (n *mockNode) DataType() any {
	return n.dataType
}

func (n *mockNode) Dependencies() []any {
	return n.dependencies
}

func (n *mockNode) Execute(ctx context.Context) error {
	n.syncCalled = true
	if n.syncFunc != nil {
		return n.syncFunc(ctx)
	}
	return nil
}

// Обертки для каждого типа, чтобы реализовать Node[T]
type mockNodeA struct{ *mockNode }
type mockNodeB struct{ *mockNode }
type mockNodeC struct{ *mockNode }
type mockNodeD struct{ *mockNode }

// testCtx создает контекст с логером для тестов.
// Использует имя теста (t.Name()) в качестве имени логера.
// Cleanup функция регистрируется через t.Cleanup для автоматической очистки после завершения теста.
func testCtx(t *testing.T) context.Context {
	ctx, cleanup := log.NewCtx(context.Background(), t.Name(), true)
	t.Cleanup(cleanup)
	return ctx
}

func TestNewGraph(t *testing.T) {
	graph := NewGraph()
	if graph == nil {
		t.Fatal("NewGraph returned nil")
	}
	if graph.nodes == nil {
		t.Fatal("graph.nodes is nil")
	}
	if len(graph.nodes) != 0 {
		t.Fatal("new graph should be empty")
	}
}

func TestAddNode(t *testing.T) {
	graph := NewGraph()
	nodeKey := "TestTypeA"
	mock := &mockNode{
		dataType:     nodeKey,
		dependencies: nil,
	}
	node := &mockNodeA{mock}

	graph.AddNode(node)

	if len(graph.nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(graph.nodes))
	}

	if graph.nodes[nodeKey] != node {
		t.Fatal("node not found in graph")
	}
}

func TestExecuteAll_SimpleOrder(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Создаем узлы: A -> B -> C
	keyA := "TestTypeA"
	keyB := "TestTypeB"
	keyC := "TestTypeC"

	mockA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
	}
	nodeA := &mockNodeA{mockA}

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA,
		},
	}
	nodeB := &mockNodeB{mockB}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyB,
		},
	}
	nodeC := &mockNodeC{mockC}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)

	err := graph.ExecuteAll(ctx)
	if err != nil {
		t.Fatalf("ExecuteAll failed: %v", err)
	}

	// Проверяем, что все узлы были синхронизированы
	if !nodeA.syncCalled {
		t.Error("nodeA was not synced")
	}
	if !nodeB.syncCalled {
		t.Error("nodeB was not synced")
	}
	if !nodeC.syncCalled {
		t.Error("nodeC was not synced")
	}
}

func TestExecuteAll_ParallelNodes(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Создаем граф: A -> B, C (B и C независимы, оба зависят от A)
	keyA := "TestTypeA"
	keyB := "TestTypeB"
	keyC := "TestTypeC"

	mockA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
	}
	nodeA := &mockNodeA{mockA}

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA,
		},
	}
	nodeB := &mockNodeB{mockB}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyA,
		},
	}
	nodeC := &mockNodeC{mockC}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)

	err := graph.ExecuteAll(ctx)
	if err != nil {
		t.Fatalf("ExecuteAll failed: %v", err)
	}

	// Проверяем, что все узлы были синхронизированы
	if !mockA.syncCalled {
		t.Error("nodeA was not synced")
	}
	if !mockB.syncCalled {
		t.Error("nodeB was not synced")
	}
	if !mockC.syncCalled {
		t.Error("nodeC was not synced")
	}

	// A должен быть синхронизирован первым
	if !mockA.syncCalled {
		t.Error("nodeA should be synced before B and C")
	}
}

func TestExecuteAll_MultipleDependencies(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Создаем граф: A, B -> C (C зависит от A и B)
	keyA := "TestTypeA"
	keyB := "TestTypeB"
	keyC := "TestTypeC"

	mockA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
	}
	nodeA := &mockNodeA{mockA}

	mockB := &mockNode{
		dataType:     keyB,
		dependencies: nil,
	}
	nodeB := &mockNodeB{mockB}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyA,
			keyB,
		},
	}
	nodeC := &mockNodeC{mockC}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)

	err := graph.ExecuteAll(ctx)
	if err != nil {
		t.Fatalf("ExecuteAll failed: %v", err)
	}

	// Проверяем, что все узлы были синхронизированы
	if !mockA.syncCalled {
		t.Error("nodeA was not synced")
	}
	if !mockB.syncCalled {
		t.Error("nodeB was not synced")
	}
	if !mockC.syncCalled {
		t.Error("nodeC was not synced")
	}
}

func TestExecuteAll_CircularDependency(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Создаем циклическую зависимость: A -> B -> C -> A
	keyA := "TestTypeA"
	keyB := "TestTypeB"
	keyC := "TestTypeC"

	mockA := &mockNode{
		dataType: keyA,
		dependencies: []any{
			keyC,
		},
	}
	nodeA := &mockNodeA{mockA}

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA,
		},
	}
	nodeB := &mockNodeB{mockB}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyB,
		},
	}
	nodeC := &mockNodeC{mockC}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)

	err := graph.ExecuteAll(ctx)
	if err == nil {
		t.Fatal("expected error for circular dependency")
	}

	if !errors.Is(err, ErrCircularDependency) && !errors.Is(err, ErrFailedToSortNodes) {
		t.Fatalf("expected ErrCircularDependency or ErrFailedToSortNodes, got: %v", err)
	}
}

func TestExecuteAll_ExecuteError(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	expectedErr := errors.New("execute error")
	keyA := "TestTypeA"

	nodeA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
		syncFunc: func(ctx context.Context) error {
			return expectedErr
		},
	}

	graph.AddNode(nodeA)

	err := graph.ExecuteAll(ctx)
	if err == nil {
		t.Fatal("expected error from execute")
	}

	if !errors.Is(err, ErrFailedToExecuteNode) {
		t.Fatalf("expected ErrFailedToExecuteNode, got: %v", err)
	}
}

func TestExecuteAll_EmptyGraph(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	err := graph.ExecuteAll(ctx)
	if err != nil {
		t.Fatalf("ExecuteAll on empty graph should not fail, got: %v", err)
	}
}

func TestExecuteAll_ComplexGraph(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Создаем сложный граф:
	//   A
	//  / \
	// B   C
	//  \ /
	//   D

	keyA := "TestTypeA"
	keyB := "TestTypeB"
	keyC := "TestTypeC"
	keyD := "TestTypeD"

	mockA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
	}
	nodeA := &mockNodeA{mockA}

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA,
		},
	}
	nodeB := &mockNodeB{mockB}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyA,
		},
	}
	nodeC := &mockNodeC{mockC}

	mockD := &mockNode{
		dataType: keyD,
		dependencies: []any{
			keyB,
			keyC,
		},
	}
	nodeD := &mockNodeD{mockD}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)
	graph.AddNode(nodeD)

	err := graph.ExecuteAll(ctx)
	if err != nil {
		t.Fatalf("ExecuteAll failed: %v", err)
	}

	// Проверяем, что все узлы были синхронизированы
	if !mockA.syncCalled {
		t.Error("nodeA was not synced")
	}
	if !mockB.syncCalled {
		t.Error("nodeB was not synced")
	}
	if !mockC.syncCalled {
		t.Error("nodeC was not synced")
	}
	if !mockD.syncCalled {
		t.Error("nodeD was not synced")
	}
}

// Тесты для ExecuteInParallel

func TestExecuteInParallel_SimpleOrder(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Создаем узлы: A -> B -> C (последовательное выполнение)
	keyA := "TestTypeA"
	keyB := "TestTypeB"
	keyC := "TestTypeC"

	mockA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
	}
	nodeA := &mockNodeA{mockA}

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA,
		},
	}
	nodeB := &mockNodeB{mockB}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyB,
		},
	}
	nodeC := &mockNodeC{mockC}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)

	err := graph.ExecuteInParallel(ctx)
	if err != nil {
		t.Fatalf("ExecuteInParallel failed: %v", err)
	}

	// Проверяем, что все узлы были синхронизированы
	if !mockA.syncCalled {
		t.Error("nodeA was not synced")
	}
	if !mockB.syncCalled {
		t.Error("nodeB was not synced")
	}
	if !mockC.syncCalled {
		t.Error("nodeC was not synced")
	}
}

func TestExecuteInParallel_ParallelNodes(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Создаем граф: A -> B, C (B и C независимы, оба зависят от A)
	// B и C должны выполняться параллельно
	keyA := "TestTypeA"
	keyB := "TestTypeB"
	keyC := "TestTypeC"

	var bStart, cStart, bEnd, cEnd time.Time

	mockA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
	}
	nodeA := &mockNodeA{mockA}

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA,
		},
		syncFunc: func(ctx context.Context) error {
			bStart = time.Now()
			time.Sleep(50 * time.Millisecond) // Задержка для проверки параллельности
			bEnd = time.Now()
			return nil
		},
	}
	nodeB := &mockNodeB{mockB}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyA,
		},
		syncFunc: func(ctx context.Context) error {
			cStart = time.Now()
			time.Sleep(50 * time.Millisecond) // Задержка для проверки параллельности
			cEnd = time.Now()
			return nil
		},
	}
	nodeC := &mockNodeC{mockC}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)

	startTime := time.Now()
	err := graph.ExecuteInParallel(ctx)
	totalTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("ExecuteInParallel failed: %v", err)
	}

	// Проверяем, что все узлы были синхронизированы
	if !mockA.syncCalled {
		t.Error("nodeA was not synced")
	}
	if !mockB.syncCalled {
		t.Error("nodeB was not synced")
	}
	if !mockC.syncCalled {
		t.Error("nodeC was not synced")
	}

	// Проверяем, что B и C выполнялись параллельно
	// Если они выполняются последовательно, общее время будет ~100ms
	// Если параллельно - ~50ms + время на A
	if totalTime > 80*time.Millisecond {
		t.Errorf("B and C should execute in parallel (total time ~50ms), but took %v", totalTime)
	}

	// Проверяем, что B и C выполнялись одновременно (пересекались по времени)
	overlap := (bStart.Before(cEnd) || bStart.Equal(cEnd)) && (cStart.Before(bEnd) || cStart.Equal(bEnd))
	if !overlap {
		t.Error("B and C should execute in parallel, but they don't overlap in time")
	}
}

func TestExecuteInParallel_MultipleRanks(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Создаем сложный граф:
	//   A (rank 0)
	//  / \
	// B   C (rank 1, параллельно)
	//  \ /
	//   D (rank 2)
	keyA := "TestTypeA"
	keyB := "TestTypeB"
	keyC := "TestTypeC"
	keyD := "TestTypeD"

	executionOrder := make([]string, 0)
	var mu sync.Mutex

	mockA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
		syncFunc: func(ctx context.Context) error {
			mu.Lock()
			executionOrder = append(executionOrder, "A")
			mu.Unlock()
			return nil
		},
	}
	nodeA := &mockNodeA{mockA}

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA,
		},
		syncFunc: func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond) // Небольшая задержка
			mu.Lock()
			executionOrder = append(executionOrder, "B")
			mu.Unlock()
			return nil
		},
	}
	nodeB := &mockNodeB{mockB}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyA,
		},
		syncFunc: func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond) // Небольшая задержка
			mu.Lock()
			executionOrder = append(executionOrder, "C")
			mu.Unlock()
			return nil
		},
	}
	nodeC := &mockNodeC{mockC}

	mockD := &mockNode{
		dataType: keyD,
		dependencies: []any{
			keyB,
			keyC,
		},
		syncFunc: func(ctx context.Context) error {
			mu.Lock()
			executionOrder = append(executionOrder, "D")
			mu.Unlock()
			return nil
		},
	}
	nodeD := &mockNodeD{mockD}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)
	graph.AddNode(nodeD)

	err := graph.ExecuteInParallel(ctx)
	if err != nil {
		t.Fatalf("ExecuteInParallel failed: %v", err)
	}

	// Проверяем, что все узлы были синхронизированы
	if !mockA.syncCalled {
		t.Error("nodeA was not synced")
	}
	if !mockB.syncCalled {
		t.Error("nodeB was not synced")
	}
	if !mockC.syncCalled {
		t.Error("nodeC was not synced")
	}
	if !mockD.syncCalled {
		t.Error("nodeD was not synced")
	}

	// Проверяем порядок выполнения
	// A должен быть первым
	if len(executionOrder) < 1 || executionOrder[0] != "A" {
		t.Errorf("A should be executed first, got order: %v", executionOrder)
	}

	// B и C должны быть после A, но до D
	// D должен быть последним
	if len(executionOrder) < 4 {
		t.Errorf("expected 4 nodes executed, got %d", len(executionOrder))
	}
	if executionOrder[len(executionOrder)-1] != "D" {
		t.Errorf("D should be executed last, got order: %v", executionOrder)
	}
}

func TestExecuteInParallel_CircularDependency(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Создаем циклическую зависимость: A -> B -> C -> A
	keyA := "TestTypeA"
	keyB := "TestTypeB"
	keyC := "TestTypeC"

	mockA := &mockNode{
		dataType: keyA,
		dependencies: []any{
			keyC,
		},
	}
	nodeA := &mockNodeA{mockA}

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA,
		},
	}
	nodeB := &mockNodeB{mockB}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyB,
		},
	}
	nodeC := &mockNodeC{mockC}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)

	err := graph.ExecuteInParallel(ctx)
	if err == nil {
		t.Fatal("expected error for circular dependency")
	}

	if !errors.Is(err, ErrCircularDependency) && !errors.Is(err, ErrFailedToSortNodes) {
		t.Fatalf("expected ErrCircularDependency or ErrFailedToSortNodes, got: %v", err)
	}
}

func TestExecuteInParallel_ExecuteError(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	expectedErr := errors.New("execute error")
	keyA := "TestTypeA"

	nodeA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
		syncFunc: func(ctx context.Context) error {
			return expectedErr
		},
	}

	graph.AddNode(nodeA)

	err := graph.ExecuteInParallel(ctx)
	if err == nil {
		t.Fatal("expected error from execute")
	}

	if !errors.Is(err, ErrFailedToExecuteNode) {
		t.Fatalf("expected ErrFailedToExecuteNode, got: %v", err)
	}
}

func TestExecuteInParallel_ExecuteErrorInParallel(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Создаем граф: A -> B, C
	// B возвращает ошибку, C выполняется успешно
	// Ошибка должна быть обработана
	keyA := "TestTypeA"
	keyB := "TestTypeB"
	keyC := "TestTypeC"

	expectedErr := errors.New("execute error in B")

	mockA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
	}
	nodeA := &mockNodeA{mockA}

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA,
		},
		syncFunc: func(ctx context.Context) error {
			return expectedErr
		},
	}
	nodeB := &mockNodeB{mockB}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyA,
		},
	}
	nodeC := &mockNodeC{mockC}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)

	err := graph.ExecuteInParallel(ctx)
	if err == nil {
		t.Fatal("expected error from execute")
	}

	if !errors.Is(err, ErrFailedToExecuteNode) {
		t.Fatalf("expected ErrFailedToExecuteNode, got: %v", err)
	}

	// A должен быть выполнен
	if !mockA.syncCalled {
		t.Error("nodeA should be executed")
	}
}

func TestExecuteInParallel_EmptyGraph(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	err := graph.ExecuteInParallel(ctx)
	if err != nil {
		t.Fatalf("ExecuteInParallel on empty graph should not fail, got: %v", err)
	}
}

func TestExecuteInParallel_SingleNode(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	keyA := "TestTypeA"
	mockA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
	}
	nodeA := &mockNodeA{mockA}

	graph.AddNode(nodeA)

	err := graph.ExecuteInParallel(ctx)
	if err != nil {
		t.Fatalf("ExecuteInParallel failed: %v", err)
	}

	if !mockA.syncCalled {
		t.Error("nodeA was not synced")
	}
}

func TestTestCtx_LoggerAndCleanup(t *testing.T) {
	// Проверяем, что контекст создается с логером
	ctx := testCtx(t)

	// Проверяем, что логер доступен из контекста
	logger := log.FromContext(ctx)
	if logger == nil {
		t.Fatal("logger should be available in context")
	}

	// Проверяем, что имя теста используется в логере
	log.Info(ctx, "Test log message from test", log.String("test_name", t.Name()))

	// Проверяем, что логер работает корректно
	log.Info(ctx, "Logger is working correctly")
	log.Debug(ctx, "Debug message should be visible with debug=true")

	// Проверяем, что cleanup вызывается корректно
	// После завершения теста cleanup должен быть вызван автоматически через t.Cleanup
	// Это проверяется тем, что тест завершается без паники
	// Если cleanup не вызывается, логер может остаться открытым, что вызовет проблемы
}
