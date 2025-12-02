package depsgraph

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

// Тестовые типы данных
type TestTypeA struct{}
type TestTypeB struct{}
type TestTypeC struct{}
type TestTypeD struct{}

// Mock узел для тестирования
type mockNode struct {
	dataType     reflect.Type
	dependencies []reflect.Type
	syncFunc     func(ctx context.Context) error
	syncCalled   bool
}

func (n *mockNode) DataType() reflect.Type {
	return n.dataType
}

func (n *mockNode) Dependencies() []reflect.Type {
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
	mock := &mockNode{
		dataType:     reflect.TypeOf((*TestTypeA)(nil)).Elem(),
		dependencies: nil,
	}
	node := &mockNodeA{mock}

	AddNode[TestTypeA](graph, node)

	if len(graph.nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(graph.nodes))
	}

	typ := reflect.TypeOf((*TestTypeA)(nil)).Elem()
	if graph.nodes[typ] != node {
		t.Fatal("node not found in graph")
	}
}

func TestExecuteAll_SimpleOrder(t *testing.T) {
	graph := NewGraph()
	ctx := context.Background()

	// Создаем узлы: A -> B -> C
	mockA := &mockNode{
		dataType:     reflect.TypeOf((*TestTypeA)(nil)).Elem(),
		dependencies: nil,
	}
	nodeA := &mockNodeA{mockA}

	mockB := &mockNode{
		dataType: reflect.TypeOf((*TestTypeB)(nil)).Elem(),
		dependencies: []reflect.Type{
			reflect.TypeOf((*TestTypeA)(nil)).Elem(),
		},
	}
	nodeB := &mockNodeB{mockB}

	mockC := &mockNode{
		dataType: reflect.TypeOf((*TestTypeC)(nil)).Elem(),
		dependencies: []reflect.Type{
			reflect.TypeOf((*TestTypeB)(nil)).Elem(),
		},
	}
	nodeC := &mockNodeC{mockC}

	AddNode[TestTypeA](graph, nodeA)
	AddNode[TestTypeB](graph, nodeB)
	AddNode[TestTypeC](graph, nodeC)

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
	ctx := context.Background()

	// Создаем граф: A -> B, C (B и C независимы, оба зависят от A)
	mockA := &mockNode{
		dataType:     reflect.TypeOf((*TestTypeA)(nil)).Elem(),
		dependencies: nil,
	}
	nodeA := &mockNodeA{mockA}

	mockB := &mockNode{
		dataType: reflect.TypeOf((*TestTypeB)(nil)).Elem(),
		dependencies: []reflect.Type{
			reflect.TypeOf((*TestTypeA)(nil)).Elem(),
		},
	}
	nodeB := &mockNodeB{mockB}

	mockC := &mockNode{
		dataType: reflect.TypeOf((*TestTypeC)(nil)).Elem(),
		dependencies: []reflect.Type{
			reflect.TypeOf((*TestTypeA)(nil)).Elem(),
		},
	}
	nodeC := &mockNodeC{mockC}

	AddNode[TestTypeA](graph, nodeA)
	AddNode[TestTypeB](graph, nodeB)
	AddNode[TestTypeC](graph, nodeC)

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
	ctx := context.Background()

	// Создаем граф: A, B -> C (C зависит от A и B)
	mockA := &mockNode{
		dataType:     reflect.TypeOf((*TestTypeA)(nil)).Elem(),
		dependencies: nil,
	}
	nodeA := &mockNodeA{mockA}

	mockB := &mockNode{
		dataType:     reflect.TypeOf((*TestTypeB)(nil)).Elem(),
		dependencies: nil,
	}
	nodeB := &mockNodeB{mockB}

	mockC := &mockNode{
		dataType: reflect.TypeOf((*TestTypeC)(nil)).Elem(),
		dependencies: []reflect.Type{
			reflect.TypeOf((*TestTypeA)(nil)).Elem(),
			reflect.TypeOf((*TestTypeB)(nil)).Elem(),
		},
	}
	nodeC := &mockNodeC{mockC}

	AddNode[TestTypeA](graph, nodeA)
	AddNode[TestTypeB](graph, nodeB)
	AddNode[TestTypeC](graph, nodeC)

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
	ctx := context.Background()

	// Создаем циклическую зависимость: A -> B -> C -> A
	mockA := &mockNode{
		dataType: reflect.TypeOf((*TestTypeA)(nil)).Elem(),
		dependencies: []reflect.Type{
			reflect.TypeOf((*TestTypeC)(nil)).Elem(),
		},
	}
	nodeA := &mockNodeA{mockA}

	mockB := &mockNode{
		dataType: reflect.TypeOf((*TestTypeB)(nil)).Elem(),
		dependencies: []reflect.Type{
			reflect.TypeOf((*TestTypeA)(nil)).Elem(),
		},
	}
	nodeB := &mockNodeB{mockB}

	mockC := &mockNode{
		dataType: reflect.TypeOf((*TestTypeC)(nil)).Elem(),
		dependencies: []reflect.Type{
			reflect.TypeOf((*TestTypeB)(nil)).Elem(),
		},
	}
	nodeC := &mockNodeC{mockC}

	AddNode[TestTypeA](graph, nodeA)
	AddNode[TestTypeB](graph, nodeB)
	AddNode[TestTypeC](graph, nodeC)

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
	ctx := context.Background()

	expectedErr := errors.New("execute error")

	nodeA := &mockNode{
		dataType:     reflect.TypeOf((*TestTypeA)(nil)).Elem(),
		dependencies: nil,
		syncFunc: func(ctx context.Context) error {
			return expectedErr
		},
	}

	AddNode[TestTypeA](graph, nodeA)

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
	ctx := context.Background()

	err := graph.ExecuteAll(ctx)
	if err != nil {
		t.Fatalf("ExecuteAll on empty graph should not fail, got: %v", err)
	}
}

func TestExecuteAll_ComplexGraph(t *testing.T) {
	graph := NewGraph()
	ctx := context.Background()

	// Создаем сложный граф:
	//   A
	//  / \
	// B   C
	//  \ /
	//   D

	mockA := &mockNode{
		dataType:     reflect.TypeOf((*TestTypeA)(nil)).Elem(),
		dependencies: nil,
	}
	nodeA := &mockNodeA{mockA}

	mockB := &mockNode{
		dataType: reflect.TypeOf((*TestTypeB)(nil)).Elem(),
		dependencies: []reflect.Type{
			reflect.TypeOf((*TestTypeA)(nil)).Elem(),
		},
	}
	nodeB := &mockNodeB{mockB}

	mockC := &mockNode{
		dataType: reflect.TypeOf((*TestTypeC)(nil)).Elem(),
		dependencies: []reflect.Type{
			reflect.TypeOf((*TestTypeA)(nil)).Elem(),
		},
	}
	nodeC := &mockNodeC{mockC}

	mockD := &mockNode{
		dataType: reflect.TypeOf((*TestTypeD)(nil)).Elem(),
		dependencies: []reflect.Type{
			reflect.TypeOf((*TestTypeB)(nil)).Elem(),
			reflect.TypeOf((*TestTypeC)(nil)).Elem(),
		},
	}
	nodeD := &mockNodeD{mockD}

	AddNode[TestTypeA](graph, nodeA)
	AddNode[TestTypeB](graph, nodeB)
	AddNode[TestTypeC](graph, nodeC)
	AddNode[TestTypeD](graph, nodeD)

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
