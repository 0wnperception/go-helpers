// pkg/depsgraph/depsgraph.go
package depsgraph

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/0wnperception/go-helpers/pkg/log"
)

var (
	ErrCircularDependency  = errors.New("circular dependency detected")
	ErrFailedToSortNodes   = errors.New("failed to sort nodes")
	ErrFailedToExecuteNode = errors.New("failed to execute node")
)

// Node - узел графа зависимостей
// T - тип данных, который представляет этот узел
type Node[T any] interface {
	// DataType возвращает тип данных узла
	DataType() reflect.Type

	// Dependencies возвращает список типов зависимостей
	Dependencies() []reflect.Type

	// Execute выполняет операцию узла
	Execute(ctx context.Context) error
}

// Graph - граф зависимостей с топологической сортировкой
type Graph struct {
	nodes map[reflect.Type]Node[any]
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[reflect.Type]Node[any]),
	}
}

// AddNode добавляет узел в граф
func AddNode[T any](g *Graph, node Node[T]) {
	var zero T
	typ := reflect.TypeOf(zero)
	g.nodes[typ] = node
}

// ExecuteAll выполняет операции всех узлов в правильном порядке
func (g *Graph) ExecuteAll(ctx context.Context) error {
	// Топологическая сортировка
	order, err := g.topologicalSort()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToSortNodes, err)
	}

	log.Info(ctx, "Execution order determined",
		log.Int("nodes_count", len(order)),
	)

	// Выполняем операции в порядке зависимостей
	for _, typ := range order {
		node := g.nodes[typ]
		log.Info(ctx, fmt.Sprintf("Executing %s", typ.String()))

		if err := node.Execute(ctx); err != nil {
			return fmt.Errorf("%w %s: %w", ErrFailedToExecuteNode, typ.String(), err)
		}
	}

	return nil
}

// topologicalSort выполняет топологическую сортировку (Kahn's algorithm)
func (g *Graph) topologicalSort() ([]reflect.Type, error) {
	// Вычисляем in-degree для каждого узла
	inDegree := make(map[reflect.Type]int)
	for typ := range g.nodes {
		inDegree[typ] = 0
	}

	for _, node := range g.nodes {
		for range node.Dependencies() {
			inDegree[node.DataType()]++
		}
	}

	// Находим узлы без зависимостей
	queue := make([]reflect.Type, 0)
	for typ, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, typ)
		}
	}

	result := make([]reflect.Type, 0, len(g.nodes))

	// Обрабатываем узлы
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Уменьшаем in-degree для узлов, зависящих от current
		for _, node := range g.nodes {
			for _, dep := range node.Dependencies() {
				if dep == current {
					inDegree[node.DataType()]--
					if inDegree[node.DataType()] == 0 {
						queue = append(queue, node.DataType())
					}
				}
			}
		}
	}

	if len(result) != len(g.nodes) {
		return nil, ErrCircularDependency
	}

	return result, nil
}
