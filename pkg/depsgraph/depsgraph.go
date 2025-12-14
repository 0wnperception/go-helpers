// Package depsgraph provides a dependency graph implementation with topological sorting.
// It allows executing operations on nodes in the correct order considering their dependencies.
//
// The main components are:
// - Node: an interface for a node with dependencies
// - NodeWithArgs: optional interface for nodes that pass arguments between each other
// - Graph: a graph of nodes with topological sorting
// - AddNode: adding a node to the graph
// - ExecuteAll: executing operations on all nodes in the correct order
// - ExecuteInParallel: executing nodes in parallel by ranks
//
// Example usage (basic nodes without arguments):
//
//	graph := depsgraph.NewGraph()
//	graph.AddNode(&NodeA{})
//	graph.AddNode(&NodeB{})
//	graph.AddNode(&NodeC{})
//	graph.AddNode(&NodeD{})
//	return graph.ExecuteAll(ctx)
//
// The dependency graph looks like this:
//
//	  A
//	 / \
//	B   D
//	 \ /
//	  C
//
// The execution order will be:
// 1. A (no dependencies, executed first)
// 2. B or D (both depend only on A, order between them is not defined)
// 3. D or B (second from the pair)
// 4. C (always executed last, depends on A, B and D)
//
// Example usage (nodes with arguments):
//
//	// NodeA implements Node (no arguments)
//	// NodeB implements NodeWithArgs (receives result from NodeA)
//	// NodeC implements NodeWithArgs (receives results from NodeA and NodeB)
//	graph := depsgraph.NewGraph()
//	graph.AddNode(&NodeA{})
//	graph.AddNode(&NodeB{})  // Will receive result from NodeA via SetArgs
//	graph.AddNode(&NodeC{})  // Will receive results from NodeA and NodeB via SetArgs
//	return graph.ExecuteAll(ctx)
//
// When a node implements NodeWithArgs:
// - Before Execute: SetArgs is called with results from all dependencies
// - After Execute: GetResult is called and the result is stored for dependent nodes
//
// This package also supports parallel execution of nodes in different ranks.
// Nodes in the same rank (without dependencies on each other) are executed in parallel.
// Argument passing works correctly in parallel mode with proper synchronization.
// pkg/depsgraph/depsgraph.go
package depsgraph

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/0wnperception/go-helpers/pkg/log"
)

var (
	ErrCircularDependency  = errors.New("circular dependency detected")
	ErrFailedToSortNodes   = errors.New("failed to sort nodes")
	ErrFailedToExecuteNode = errors.New("failed to execute node")
	ErrInvalidDependency   = errors.New("invalid dependency: node with args depends on node without result")
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

// NodeWithArgs - опциональный интерфейс для нод с передачей аргументов.
// Если нода реализует этот интерфейс, граф автоматически передает результаты
// зависимых нод через SetArgs перед выполнением Execute.
// После выполнения Execute результат сохраняется через GetResult для передачи зависимым нодам.
type NodeWithArgs interface {
	Node

	// GetResult возвращает результат выполнения ноды.
	// Результат будет передан всем зависимым нодам через SetArgs.
	// Вызывается после успешного выполнения Execute.
	GetResult() (any, error)

	// SetArgs устанавливает аргументы от зависимых нод.
	// args - карта, где ключ - тип данных зависимости (DataType), значение - результат (GetResult).
	// Вызывается перед Execute для всех зависимостей.
	SetArgs(args map[any]any) error
}

// Graph - граф зависимостей с топологической сортировкой
type Graph struct {
	nodes map[any]Node
	// results хранит результаты выполнения нод для передачи аргументов
	// Используется только если ноды реализуют NodeWithArgs
	results map[any]any
	// resultsMu защищает доступ к results в параллельном режиме
	resultsMu sync.RWMutex
}

func NewGraph() *Graph {
	return &Graph{
		nodes:     make(map[any]Node),
		results:   make(map[any]any),
		resultsMu: sync.RWMutex{},
	}
}

// AddNode добавляет узел в граф
func (g *Graph) AddNode(node Node) {
	g.nodes[node.DataType()] = node
}

// SetInitialResult инициализирует результат для внешней зависимости.
// Это позволяет узлам иметь зависимости, которые не являются узлами графа,
// но результаты которых предоставлены извне. Такие зависимости не учитываются
// при топологической сортировке (не увеличивают in-degree узла).
// Должно вызываться до ExecuteAll или ExecuteInParallel.
func (g *Graph) SetInitialResult(depType any, result any) {
	g.resultsMu.Lock()
	defer g.resultsMu.Unlock()
	g.results[depType] = result
}

// ExecuteAll выполняет операции всех узлов в правильном порядке
func (g *Graph) ExecuteAll(ctx context.Context) error {
	// Сохраняем инициализированные результаты (внешние зависимости)
	// перед очисткой результатов выполнения узлов
	g.resultsMu.Lock()
	initialResults := make(map[any]any)
	for depType, result := range g.results {
		// Сохраняем только результаты, которые не являются узлами графа
		// (инициализированные внешние зависимости)
		if _, exists := g.nodes[depType]; !exists {
			initialResults[depType] = result
		}
	}
	// Очищаем результаты предыдущего выполнения
	g.results = make(map[any]any)
	// Восстанавливаем инициализированные результаты
	for depType, result := range initialResults {
		g.results[depType] = result
	}
	g.resultsMu.Unlock()

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
		if err := g.executeNode(ctx, typeKey); err != nil {
			return err
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

	g.resultsMu.RLock()
	for _, node := range g.nodes {
		for _, dep := range node.Dependencies() {
			// Если зависимость инициализирована (есть в results), она не учитывается
			// при подсчете in-degree, так как результат уже доступен
			if _, initialized := g.results[dep]; !initialized {
				// Зависимость не инициализирована - учитываем только если она в графе
				if _, exists := g.nodes[dep]; exists {
					inDegree[node.DataType()]++
				}
			}
		}
	}
	g.resultsMu.RUnlock()

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
					// Учитываем только если эта зависимость влияла на in-degree
					// (т.е. она была в графе и не была инициализирована)
					g.resultsMu.RLock()
					_, initialized := g.results[dep]
					g.resultsMu.RUnlock()
					if !initialized {
						inDegree[typeKey]--
						if inDegree[typeKey] == 0 {
							queue = append(queue, typeKey)
						}
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

// executeNode выполняет одну ноду с поддержкой передачи аргументов.
// Если нода реализует NodeWithArgs, перед выполнением устанавливаются аргументы
// от зависимых нод, а после выполнения сохраняется результат.
func (g *Graph) executeNode(ctx context.Context, typeKey any) error {
	node := g.nodes[typeKey]
	typeStr := fmt.Sprintf("%v", typeKey)

	// Проверяем, реализует ли нода интерфейс NodeWithArgs
	nodeWithArgs, hasArgs := node.(NodeWithArgs)

	// Если нода поддерживает аргументы, устанавливаем их перед выполнением
	if hasArgs {
		args, err := g.collectDependencyArgs(node)
		if err != nil {
			return fmt.Errorf("failed to collect args for %s: %w", typeStr, err)
		}

		if err := nodeWithArgs.SetArgs(args); err != nil {
			return fmt.Errorf("failed to set args for %s: %w", typeStr, err)
		}
	}

	// Выполняем ноду
	if err := node.Execute(ctx); err != nil {
		return fmt.Errorf("%w %s: %w", ErrFailedToExecuteNode, typeStr, err)
	}

	// Если нода поддерживает аргументы, сохраняем результат
	if hasArgs {
		result, err := nodeWithArgs.GetResult()
		if err != nil {
			return fmt.Errorf("failed to get result from %s: %w", typeStr, err)
		}

		g.resultsMu.Lock()
		g.results[typeKey] = result
		g.resultsMu.Unlock()
	}

	return nil
}

// collectDependencyArgs собирает аргументы от всех зависимых нод.
// Используется для передачи результатов зависимых нод через SetArgs.
func (g *Graph) collectDependencyArgs(node Node) (map[any]any, error) {
	deps := node.Dependencies()
	if len(deps) == 0 {
		return make(map[any]any), nil
	}

	args := make(map[any]any, len(deps))

	g.resultsMu.RLock()
	defer g.resultsMu.RUnlock()

	for _, depType := range deps {
		// Получаем результат зависимости
		result, exists := g.results[depType]
		if !exists {
			// Проверяем, существует ли зависимость в графе
			if _, nodeExists := g.nodes[depType]; nodeExists {
				// Зависимость в графе, но результат недоступен - это означает невалидную конфигурацию:
				// 1. Зависимость еще не выполнена (не должно происходить при правильном порядке)
				// 2. Зависимость не реализует NodeWithArgs (невалидная конфигурация)
				// Нода с аргументами не может зависеть от ноды без результата
				return nil, fmt.Errorf("%w: dependency %v does not provide result (node with args depends on node without NodeWithArgs or dependency not executed)", ErrInvalidDependency, depType)
			}
			// Зависимость не в графе и не инициализирована - пропускаем
			// (внешняя зависимость без инициализации)
			continue
		}

		args[depType] = result
	}

	return args, nil
}

// getRanks группирует ноды по рангам (уровням).
// Ранг 0: ноды без зависимостей
// Ранг 1: ноды, зависящие только от нод ранга 0
// И так далее.
func (g *Graph) getRanks() ([][]any, error) {
	// Вычисляем in-degree для каждого узла
	inDegree := make(map[any]int)
	for typeKey := range g.nodes {
		inDegree[typeKey] = 0
	}

	// Подсчитываем зависимости
	g.resultsMu.RLock()
	for _, node := range g.nodes {
		for _, dep := range node.Dependencies() {
			// Если зависимость инициализирована (есть в results), она не учитывается
			// при подсчете in-degree, так как результат уже доступен
			if _, initialized := g.results[dep]; !initialized {
				// Зависимость не инициализирована - учитываем только если она в графе
				if _, exists := g.nodes[dep]; exists {
					inDegree[node.DataType()]++
				}
			}
		}
	}
	g.resultsMu.RUnlock()

	ranks := make([][]any, 0)
	processed := make(map[any]bool)

	// Пока есть необработанные ноды
	for len(processed) < len(g.nodes) {
		// Находим ноды текущего ранга (in-degree = 0 и еще не обработаны)
		currentRank := make([]any, 0)
		for typeKey, degree := range inDegree {
			if degree == 0 && !processed[typeKey] {
				currentRank = append(currentRank, typeKey)
				processed[typeKey] = true
			}
		}

		if len(currentRank) == 0 {
			// Если не нашли нод для обработки, но есть необработанные - циклическая зависимость
			return nil, ErrCircularDependency
		}

		ranks = append(ranks, currentRank)

		// Уменьшаем in-degree для нод, зависящих от текущего ранга
		g.resultsMu.RLock()
		for _, typeKey := range currentRank {
			for _, node := range g.nodes {
				for _, dep := range node.Dependencies() {
					if dep == typeKey {
						nodeTypeKey := node.DataType()
						if !processed[nodeTypeKey] {
							// Учитываем только если эта зависимость влияла на in-degree
							// (т.е. она была в графе и не была инициализирована)
							_, initialized := g.results[dep]
							if !initialized {
								inDegree[nodeTypeKey]--
							}
						}
					}
				}
			}
		}
		g.resultsMu.RUnlock()
	}

	return ranks, nil
}

// ExecuteInParallel выполняет операции всех узлов параллельно по рангам.
// Ноды одного ранга (без зависимостей друг от друга) выполняются параллельно.
// Если в ранге только одна нода, выполняется последовательно.
func (g *Graph) ExecuteInParallel(ctx context.Context) error {
	// Сохраняем инициализированные результаты (внешние зависимости)
	// перед очисткой результатов выполнения узлов
	g.resultsMu.Lock()
	initialResults := make(map[any]any)
	for depType, result := range g.results {
		// Сохраняем только результаты, которые не являются узлами графа
		// (инициализированные внешние зависимости)
		if _, exists := g.nodes[depType]; !exists {
			initialResults[depType] = result
		}
	}
	// Очищаем результаты предыдущего выполнения
	g.results = make(map[any]any)
	// Восстанавливаем инициализированные результаты
	for depType, result := range initialResults {
		g.results[depType] = result
	}
	g.resultsMu.Unlock()

	// Получаем ранги нод
	ranks, err := g.getRanks()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToSortNodes, err)
	}

	log.Info(ctx, "Execution ranks determined",
		log.Int("ranks_count", len(ranks)),
	)

	// Рекурсивно выполняем ноды по рангам
	return g.executeRank(ctx, ranks, 0)
}

// executeRank рекурсивно выполняет ноды начиная с указанного ранга.
func (g *Graph) executeRank(ctx context.Context, ranks [][]any, rankIndex int) error {
	// Базовый случай: все ранги обработаны
	if rankIndex >= len(ranks) {
		return nil
	}

	currentRank := ranks[rankIndex]
	if len(currentRank) == 0 {
		// Пустой ранг - переходим к следующему
		return g.executeRank(ctx, ranks, rankIndex+1)
	}

	// Если в ранге только одна нода, выполняем последовательно
	if len(currentRank) == 1 {
		typeKey := currentRank[0]
		typeStr := fmt.Sprintf("%v", typeKey)
		log.Info(ctx, fmt.Sprintf("Executing %s (rank %d)", typeStr, rankIndex))

		if err := g.executeNode(ctx, typeKey); err != nil {
			return err
		}

		// Рекурсивно вызываем для следующего ранга
		return g.executeRank(ctx, ranks, rankIndex+1)
	}

	// Если в ранге несколько нод, выполняем параллельно
	log.Info(ctx, fmt.Sprintf("Executing rank %d with %d nodes in parallel", rankIndex, len(currentRank)))

	var wg sync.WaitGroup
	errChan := make(chan error, len(currentRank))

	for _, typeKey := range currentRank {
		wg.Add(1)
		go func(key any) {
			defer wg.Done()

			typeStr := fmt.Sprintf("%v", key)
			log.Info(ctx, fmt.Sprintf("Executing %s (rank %d)", typeStr, rankIndex))

			if err := g.executeNode(ctx, key); err != nil {
				errChan <- err
				return
			}
		}(typeKey)
	}

	wg.Wait()
	close(errChan)

	// Проверяем ошибки
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	// Рекурсивно вызываем для следующего ранга
	return g.executeRank(ctx, ranks, rankIndex+1)
}
