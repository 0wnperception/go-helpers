// pkg/depsgraph/depsgraph.go
package depsgraph

import (
	"context"
	"errors"
	"fmt"

	"github.com/0wnperception/go-helpers/pkg/log"
)

var (
	ErrCircularDependency  = errors.New("circular dependency detected")
	ErrFailedToSortNodes   = errors.New("failed to sort nodes")
	ErrFailedToExecuteNode = errors.New("failed to execute node")
)

// Node - узел графа зависимостей
type Node interface {
	// DataType возвращает тип данных узла
	DataType() any

	// Dependencies возвращает список типов зависимостей
	Dependencies() []any

	// Execute выполняет операцию узла
	Execute(ctx context.Context) error
}

// Graph - граф зависимостей с топологической сортировкой
type Graph struct {
	nodes map[any]Node
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[any]Node),
	}
}

// AddNode добавляет узел в граф
func (g *Graph) AddNode(node Node) {
	g.nodes[node.DataType()] = node
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
	for _, typeKey := range order {
		node := g.nodes[typeKey]
		typeStr := fmt.Sprintf("%v", typeKey)
		log.Info(ctx, fmt.Sprintf("Executing %s", typeStr))

		if err := node.Execute(ctx); err != nil {
			return fmt.Errorf("%w %s: %w", ErrFailedToExecuteNode, typeStr, err)
		}
	}

	return nil
}

// topologicalSort выполняет топологическую сортировку (Kahn's algorithm)
func (g *Graph) topologicalSort() ([]any, error) {
	// Вычисляем in-degree для каждого узла
	inDegree := make(map[any]int)
	for typeKey := range g.nodes {
		inDegree[typeKey] = 0
	}

	for _, node := range g.nodes {
		for range node.Dependencies() {
			inDegree[node.DataType()]++
		}
	}

	// Находим узлы без зависимостей
	queue := make([]any, 0)
	for typeKey, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, typeKey)
		}
	}

	result := make([]any, 0, len(g.nodes))

	// Обрабатываем узлы
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Уменьшаем in-degree для узлов, зависящих от current
		for _, node := range g.nodes {
			for _, dep := range node.Dependencies() {
				if dep == current {
					typeKey := node.DataType()
					inDegree[typeKey]--
					if inDegree[typeKey] == 0 {
						queue = append(queue, typeKey)
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
