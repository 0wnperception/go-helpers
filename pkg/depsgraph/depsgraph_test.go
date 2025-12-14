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

// Тесты для NodeWithArgs

// mockNodeWithArgs - mock нода с поддержкой передачи аргументов
type mockNodeWithArgs struct {
	*mockNode
	result      any
	args        map[any]any
	setArgsFunc func(args map[any]any) error
}

func (n *mockNodeWithArgs) GetResult() (any, error) {
	return n.result, nil
}

func (n *mockNodeWithArgs) SetArgs(args map[any]any) error {
	n.args = args
	if n.setArgsFunc != nil {
		return n.setArgsFunc(args)
	}
	return nil
}

func TestExecuteAll_WithArgs_SimplePipeline(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Создаем конвейер: A -> B -> C
	// A возвращает результат, B получает его и возвращает свой, C получает результат B
	keyA := "TestTypeA"
	keyB := "TestTypeB"
	keyC := "TestTypeC"

	resultA := "resultA"
	resultB := "resultB"

	mockA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
	}
	nodeA := &mockNodeWithArgs{
		mockNode: mockA,
		result:   resultA,
	}

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA,
		},
	}
	nodeB := &mockNodeWithArgs{
		mockNode: mockB,
		result:   resultB,
	}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyB,
		},
	}
	nodeC := &mockNodeWithArgs{
		mockNode: mockC,
		result:   "resultC",
	}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)

	err := graph.ExecuteAll(ctx)
	if err != nil {
		t.Fatalf("ExecuteAll failed: %v", err)
	}

	// Проверяем, что все ноды были выполнены
	if !mockA.syncCalled {
		t.Error("nodeA was not executed")
	}
	if !mockB.syncCalled {
		t.Error("nodeB was not executed")
	}
	if !mockC.syncCalled {
		t.Error("nodeC was not executed")
	}

	// Проверяем, что аргументы были переданы
	if nodeB.args == nil {
		t.Error("nodeB should have received args")
	}
	if nodeB.args[keyA] != resultA {
		t.Errorf("nodeB should have received resultA, got %v", nodeB.args[keyA])
	}

	if nodeC.args == nil {
		t.Error("nodeC should have received args")
	}
	if nodeC.args[keyB] != resultB {
		t.Errorf("nodeC should have received resultB, got %v", nodeC.args[keyB])
	}
}

func TestExecuteAll_WithArgs_MultipleDependencies(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Создаем граф: A, B -> C (C получает результаты от A и B)
	keyA := "TestTypeA"
	keyB := "TestTypeB"
	keyC := "TestTypeC"

	resultA := "resultA"
	resultB := "resultB"

	mockA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
	}
	nodeA := &mockNodeWithArgs{
		mockNode: mockA,
		result:   resultA,
	}

	mockB := &mockNode{
		dataType:     keyB,
		dependencies: nil,
	}
	nodeB := &mockNodeWithArgs{
		mockNode: mockB,
		result:   resultB,
	}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyA,
			keyB,
		},
	}
	nodeC := &mockNodeWithArgs{
		mockNode: mockC,
		result:   "resultC",
	}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)

	err := graph.ExecuteAll(ctx)
	if err != nil {
		t.Fatalf("ExecuteAll failed: %v", err)
	}

	// Проверяем, что C получил аргументы от A и B
	if nodeC.args == nil {
		t.Error("nodeC should have received args")
	}
	if nodeC.args[keyA] != resultA {
		t.Errorf("nodeC should have received resultA, got %v", nodeC.args[keyA])
	}
	if nodeC.args[keyB] != resultB {
		t.Errorf("nodeC should have received resultB, got %v", nodeC.args[keyB])
	}
	if len(nodeC.args) != 2 {
		t.Errorf("nodeC should have received 2 args, got %d", len(nodeC.args))
	}
}

func TestExecuteAll_WithArgs_InvalidDependency(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Создаем невалидный граф: A (без args) -> B (с args)
	// B требует аргументы от A, но A не реализует NodeWithArgs
	// Это должно вызвать ошибку ErrInvalidDependency
	keyA := "TestTypeA"
	keyB := "TestTypeB"

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
	nodeB := &mockNodeWithArgs{
		mockNode: mockB,
		result:   "resultB",
	}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)

	err := graph.ExecuteAll(ctx)
	if err == nil {
		t.Fatal("expected error for invalid dependency configuration")
	}

	// Проверяем, что ошибка содержит информацию о невалидной зависимости
	if !errors.Is(err, ErrInvalidDependency) {
		// Проверяем, что ошибка содержит информацию о невалидной зависимости
		errStr := err.Error()
		if !contains(errStr, "invalid dependency") && !contains(errStr, "does not provide result") {
			t.Fatalf("expected ErrInvalidDependency, got: %v", err)
		}
	}

	// Проверяем, что A был выполнен (выполнение останавливается на B)
	if !mockA.syncCalled {
		t.Error("nodeA should be executed before error")
	}
}

func TestExecuteAll_WithArgs_SetArgsError(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	keyA := "TestTypeA"
	keyB := "TestTypeB"

	expectedErr := errors.New("setArgs error")

	mockA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
	}
	nodeA := &mockNodeWithArgs{
		mockNode: mockA,
		result:   "resultA",
	}

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA,
		},
	}
	nodeB := &mockNodeWithArgs{
		mockNode: mockB,
		setArgsFunc: func(args map[any]any) error {
			return expectedErr
		},
	}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)

	err := graph.ExecuteAll(ctx)
	if err == nil {
		t.Fatal("expected error from SetArgs")
	}

	if !errors.Is(err, expectedErr) && !errors.Is(err, ErrFailedToExecuteNode) {
		t.Fatalf("expected SetArgs error, got: %v", err)
	}
}

func TestExecuteAll_WithArgs_GetResultError(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	keyA := "TestTypeA"
	keyB := "TestTypeB"

	expectedErr := errors.New("getResult error")

	mockA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
	}
	nodeA := &mockNodeWithArgs{
		mockNode: mockA,
		result:   "resultA",
	}

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA,
		},
	}
	// Создаем ноду с ошибкой в GetResult
	nodeB := &mockNodeWithArgsError{
		mockNode:     mockB,
		result:       nil,
		getResultErr: expectedErr,
	}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)

	err := graph.ExecuteAll(ctx)
	if err == nil {
		t.Fatal("expected error from GetResult")
	}

	// Проверяем, что ошибка содержит информацию о GetResult
	if !errors.Is(err, expectedErr) {
		// Проверяем, что ошибка содержит информацию о GetResult
		errStr := err.Error()
		if !contains(errStr, "get result") && !contains(errStr, "GetResult") {
			t.Fatalf("expected GetResult error, got: %v", err)
		}
	}
}

// mockNodeWithArgsError - mock нода с ошибкой в GetResult
type mockNodeWithArgsError struct {
	*mockNode
	result       any
	args         map[any]any
	setArgsFunc  func(args map[any]any) error
	getResultErr error
}

func (n *mockNodeWithArgsError) GetResult() (any, error) {
	return n.result, n.getResultErr
}

func (n *mockNodeWithArgsError) SetArgs(args map[any]any) error {
	n.args = args
	if n.setArgsFunc != nil {
		return n.setArgsFunc(args)
	}
	return nil
}

// contains проверяет, содержит ли строка подстроку (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestExecuteInParallel_WithArgs_SimplePipeline(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Создаем конвейер: A -> B -> C
	keyA := "TestTypeA"
	keyB := "TestTypeB"
	keyC := "TestTypeC"

	resultA := "resultA"
	resultB := "resultB"

	mockA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
	}
	nodeA := &mockNodeWithArgs{
		mockNode: mockA,
		result:   resultA,
	}

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA,
		},
	}
	nodeB := &mockNodeWithArgs{
		mockNode: mockB,
		result:   resultB,
	}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyB,
		},
	}
	nodeC := &mockNodeWithArgs{
		mockNode: mockC,
		result:   "resultC",
	}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)

	err := graph.ExecuteInParallel(ctx)
	if err != nil {
		t.Fatalf("ExecuteInParallel failed: %v", err)
	}

	// Проверяем, что все ноды были выполнены
	if !mockA.syncCalled {
		t.Error("nodeA was not executed")
	}
	if !mockB.syncCalled {
		t.Error("nodeB was not executed")
	}
	if !mockC.syncCalled {
		t.Error("nodeC was not executed")
	}

	// Проверяем, что аргументы были переданы
	if nodeB.args == nil {
		t.Error("nodeB should have received args")
	}
	if nodeB.args[keyA] != resultA {
		t.Errorf("nodeB should have received resultA, got %v", nodeB.args[keyA])
	}

	if nodeC.args == nil {
		t.Error("nodeC should have received args")
	}
	if nodeC.args[keyB] != resultB {
		t.Errorf("nodeC should have received resultB, got %v", nodeC.args[keyB])
	}
}

func TestExecuteInParallel_WithArgs_MultipleDependencies(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Создаем граф: A -> B, C (B и C параллельно получают результат от A)
	keyA := "TestTypeA"
	keyB := "TestTypeB"
	keyC := "TestTypeC"

	resultA := "resultA"

	mockA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
	}
	nodeA := &mockNodeWithArgs{
		mockNode: mockA,
		result:   resultA,
	}

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA,
		},
	}
	nodeB := &mockNodeWithArgs{
		mockNode: mockB,
		result:   "resultB",
	}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyA,
		},
	}
	nodeC := &mockNodeWithArgs{
		mockNode: mockC,
		result:   "resultC",
	}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)

	err := graph.ExecuteInParallel(ctx)
	if err != nil {
		t.Fatalf("ExecuteInParallel failed: %v", err)
	}

	// Проверяем, что все ноды были выполнены
	if !mockA.syncCalled {
		t.Error("nodeA was not executed")
	}
	if !mockB.syncCalled {
		t.Error("nodeB was not executed")
	}
	if !mockC.syncCalled {
		t.Error("nodeC was not executed")
	}

	// Проверяем, что B и C получили аргументы от A
	if nodeB.args == nil {
		t.Error("nodeB should have received args")
	}
	if nodeB.args[keyA] != resultA {
		t.Errorf("nodeB should have received resultA, got %v", nodeB.args[keyA])
	}

	if nodeC.args == nil {
		t.Error("nodeC should have received args")
	}
	if nodeC.args[keyA] != resultA {
		t.Errorf("nodeC should have received resultA, got %v", nodeC.args[keyA])
	}
}

func TestExecuteInParallel_WithArgs_ComplexGraph(t *testing.T) {
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

	resultA := "resultA"
	resultB := "resultB"
	resultC := "resultC"

	mockA := &mockNode{
		dataType:     keyA,
		dependencies: nil,
	}
	nodeA := &mockNodeWithArgs{
		mockNode: mockA,
		result:   resultA,
	}

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA,
		},
	}
	nodeB := &mockNodeWithArgs{
		mockNode: mockB,
		result:   resultB,
	}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyA,
		},
	}
	nodeC := &mockNodeWithArgs{
		mockNode: mockC,
		result:   resultC,
	}

	mockD := &mockNode{
		dataType: keyD,
		dependencies: []any{
			keyB,
			keyC,
		},
	}
	nodeD := &mockNodeWithArgs{
		mockNode: mockD,
		result:   "resultD",
	}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)
	graph.AddNode(nodeD)

	err := graph.ExecuteInParallel(ctx)
	if err != nil {
		t.Fatalf("ExecuteInParallel failed: %v", err)
	}

	// Проверяем, что все ноды были выполнены
	if !mockA.syncCalled {
		t.Error("nodeA was not executed")
	}
	if !mockB.syncCalled {
		t.Error("nodeB was not executed")
	}
	if !mockC.syncCalled {
		t.Error("nodeC was not executed")
	}
	if !mockD.syncCalled {
		t.Error("nodeD was not executed")
	}

	// Проверяем, что аргументы были переданы корректно
	if nodeB.args[keyA] != resultA {
		t.Errorf("nodeB should have received resultA, got %v", nodeB.args[keyA])
	}
	if nodeC.args[keyA] != resultA {
		t.Errorf("nodeC should have received resultA, got %v", nodeC.args[keyA])
	}
	if nodeD.args[keyB] != resultB {
		t.Errorf("nodeD should have received resultB, got %v", nodeD.args[keyB])
	}
	if nodeD.args[keyC] != resultC {
		t.Errorf("nodeD should have received resultC, got %v", nodeD.args[keyC])
	}
}

func TestSetInitialResult_ExternalDependency(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Создаем граф, где первый узел (B) зависит от внешней зависимости (A)
	// A не является узлом графа, но результат инициализирован через SetInitialResult
	keyA := "TestTypeA"
	keyB := "TestTypeB"
	keyC := "TestTypeC"

	externalResultA := "externalResultA"
	resultB := "resultB"

	// Инициализируем результат для внешней зависимости
	graph.SetInitialResult(keyA, externalResultA)

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA, // Зависимость от внешнего узла
		},
	}
	nodeB := &mockNodeWithArgs{
		mockNode: mockB,
		result:   resultB,
	}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyB,
		},
	}
	nodeC := &mockNodeWithArgs{
		mockNode: mockC,
		result:   "resultC",
	}

	graph.AddNode(nodeB)
	graph.AddNode(nodeC)

	err := graph.ExecuteAll(ctx)
	if err != nil {
		t.Fatalf("ExecuteAll failed: %v", err)
	}

	// Проверяем, что все ноды были выполнены
	if !mockB.syncCalled {
		t.Error("nodeB was not executed")
	}
	if !mockC.syncCalled {
		t.Error("nodeC was not executed")
	}

	// Проверяем, что B получил аргументы от инициализированной внешней зависимости
	if nodeB.args == nil {
		t.Error("nodeB should have received args")
	}
	if nodeB.args[keyA] != externalResultA {
		t.Errorf("nodeB should have received externalResultA, got %v", nodeB.args[keyA])
	}

	// Проверяем, что C получил аргументы от B
	if nodeC.args == nil {
		t.Error("nodeC should have received args")
	}
	if nodeC.args[keyB] != resultB {
		t.Errorf("nodeC should have received resultB, got %v", nodeC.args[keyB])
	}
}

func TestSetInitialResult_SortingNotBroken(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Проверяем, что узел с инициализированной зависимостью может быть в первом ранге
	// (не должен иметь in-degree > 0)
	keyA := "TestTypeA"
	keyB := "TestTypeB"
	keyC := "TestTypeC"

	externalResultA := "externalResultA"

	// Инициализируем результат для внешней зависимости
	graph.SetInitialResult(keyA, externalResultA)

	executionOrder := make([]string, 0)
	var mu sync.Mutex

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA, // Зависимость от внешнего узла (инициализирована)
		},
		syncFunc: func(ctx context.Context) error {
			mu.Lock()
			executionOrder = append(executionOrder, "B")
			mu.Unlock()
			return nil
		},
	}
	nodeB := &mockNodeWithArgs{
		mockNode: mockB,
		result:   "resultB",
	}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyB,
		},
		syncFunc: func(ctx context.Context) error {
			mu.Lock()
			executionOrder = append(executionOrder, "C")
			mu.Unlock()
			return nil
		},
	}
	nodeC := &mockNodeWithArgs{
		mockNode: mockC,
		result:   "resultC",
	}

	graph.AddNode(nodeB)
	graph.AddNode(nodeC)

	err := graph.ExecuteAll(ctx)
	if err != nil {
		t.Fatalf("ExecuteAll failed: %v", err)
	}

	// Проверяем порядок выполнения
	// B должен быть первым (так как его зависимость инициализирована, in-degree = 0)
	if len(executionOrder) < 2 {
		t.Fatalf("expected 2 nodes executed, got %d", len(executionOrder))
	}
	if executionOrder[0] != "B" {
		t.Errorf("B should be executed first (has initialized dependency), got order: %v", executionOrder)
	}
	if executionOrder[1] != "C" {
		t.Errorf("C should be executed second, got order: %v", executionOrder)
	}
}

func TestSetInitialResult_ExecuteInParallel(t *testing.T) {
	graph := NewGraph()
	ctx := testCtx(t)

	// Тестируем SetInitialResult с ExecuteInParallel
	keyA := "TestTypeA"
	keyB := "TestTypeB"
	keyC := "TestTypeC"

	externalResultA := "externalResultA"
	resultB := "resultB"

	// Инициализируем результат для внешней зависимости
	graph.SetInitialResult(keyA, externalResultA)

	mockB := &mockNode{
		dataType: keyB,
		dependencies: []any{
			keyA,
		},
	}
	nodeB := &mockNodeWithArgs{
		mockNode: mockB,
		result:   resultB,
	}

	mockC := &mockNode{
		dataType: keyC,
		dependencies: []any{
			keyA, // Оба B и C зависят от A, должны выполняться параллельно
		},
	}
	nodeC := &mockNodeWithArgs{
		mockNode: mockC,
		result:   "resultC",
	}

	graph.AddNode(nodeB)
	graph.AddNode(nodeC)

	err := graph.ExecuteInParallel(ctx)
	if err != nil {
		t.Fatalf("ExecuteInParallel failed: %v", err)
	}

	// Проверяем, что все ноды были выполнены
	if !mockB.syncCalled {
		t.Error("nodeB was not executed")
	}
	if !mockC.syncCalled {
		t.Error("nodeC was not executed")
	}

	// Проверяем, что B и C получили аргументы от инициализированной зависимости
	if nodeB.args[keyA] != externalResultA {
		t.Errorf("nodeB should have received externalResultA, got %v", nodeB.args[keyA])
	}
	if nodeC.args[keyA] != externalResultA {
		t.Errorf("nodeC should have received externalResultA, got %v", nodeC.args[keyA])
	}
}
