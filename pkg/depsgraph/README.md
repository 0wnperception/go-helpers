# depsgraph

Пакет для управления графом зависимостей с топологической сортировкой. Позволяет выполнять операции над узлами в правильном порядке с учетом их зависимостей. Поддерживает передачу аргументов между узлами и параллельное выполнение.

## Основные компоненты

- `Node` - интерфейс узла с зависимостями
- `NodeWithArgs` - опциональный интерфейс для узлов с передачей аргументов
- `Graph` - граф узлов с топологической сортировкой
- `AddNode` - добавление узла в граф
- `ExecuteAll` - выполнение операций всех узлов последовательно в правильном порядке
- `ExecuteInParallel` - выполнение узлов параллельно по рангам (узлы одного ранга выполняются параллельно)

## Пример использования

```go
package main

import (
    "context"
    
    "github.com/0wnperception/go-helpers/pkg/depsgraph"
)

// Определяем ключи для узлов
const (
    KeyDataA = "DataA"
    KeyDataB = "DataB"
    KeyDataC = "DataC"
    KeyDataD = "DataD"
)

// Реализуем узел без зависимостей
type NodeA struct{}

func (n *NodeA) DataType() any {
    return KeyDataA
}

func (n *NodeA) Dependencies() []any {
    return nil // нет зависимостей
}

func (n *NodeA) Execute(ctx context.Context) error {
    // выполнение операций для узла A
    return nil
}

// Реализуем узел с одной зависимостью
type NodeB struct{}

func (n *NodeB) DataType() any {
    return KeyDataB
}

func (n *NodeB) Dependencies() []any {
    return []any{
        KeyDataA, // зависит от DataA
    }
}

func (n *NodeB) Execute(ctx context.Context) error {
    // выполнение операций для узла B
    return nil
}

// Реализуем узел с несколькими зависимостями
type NodeC struct{}

func (n *NodeC) DataType() any {
    return KeyDataC
}

func (n *NodeC) Dependencies() []any {
    return []any{
        KeyDataA, // зависит от DataA
        KeyDataB, // зависит от DataB
    }
}

func (n *NodeC) Execute(ctx context.Context) error {
    // выполнение операций для узла C
    return nil
}

// Реализуем независимый узел
type NodeD struct{}

func (n *NodeD) DataType() any {
    return KeyDataD
}

func (n *NodeD) Dependencies() []any {
    return []any{
        KeyDataA, // зависит от DataA
    }
}

func (n *NodeD) Execute(ctx context.Context) error {
    // выполнение операций для узла D
    return nil
}

func ProcessAll(ctx context.Context) error {
    // Создаем граф
    graph := depsgraph.NewGraph()

    // Добавляем узлы в произвольном порядке
    graph.AddNode(&NodeA{})
    graph.AddNode(&NodeB{})
    graph.AddNode(&NodeC{})
    graph.AddNode(&NodeD{})

    // Выполняем операции последовательно в правильном порядке
    // Порядок будет: A -> (B или D) -> (D или B) -> C
    // (B и D могут выполняться в любом порядке после A, так как оба зависят только от A)
    // C всегда выполняется последним, так как зависит от A, B и D
    return graph.ExecuteAll(ctx)
}
```

## Структура графа зависимостей

В примере выше граф зависимостей выглядит так:

```
    A
   / \
  B   D
   \ /
    C
```

Порядок выполнения (последовательно):
1. `A` (нет зависимостей, выполняется первым)
2. `B` или `D` (оба зависят только от A, порядок между ними не определен)
3. `D` или `B` (второй из пары)
4. `C` (зависит от A, B и D, выполняется последним)

**Важно:** При использовании `ExecuteAll` узлы выполняются последовательно (не параллельно). Порядок между узлами с одинаковыми зависимостями (B и D) не гарантирован и может варьироваться между запусками.

Для параллельного выполнения используйте `ExecuteInParallel`, который выполняет узлы одного ранга параллельно.

## Передача аргументов между узлами

Для передачи результатов выполнения между узлами используйте интерфейс `NodeWithArgs`. Узлы, реализующие этот интерфейс, автоматически получают результаты зависимых узлов через `SetArgs` перед выполнением и передают свой результат через `GetResult` после выполнения.

### Пример с передачей аргументов

```go
// Узел, который возвращает результат
type NodeA struct {
    result string
}

func (n *NodeA) DataType() any {
    return KeyDataA
}

func (n *NodeA) Dependencies() []any {
    return nil
}

func (n *NodeA) Execute(ctx context.Context) error {
    n.result = "result from A"
    return nil
}

func (n *NodeA) GetResult() (any, error) {
    return n.result, nil
}

func (n *NodeA) SetArgs(args map[any]any) error {
    // Нет зависимостей, аргументы не нужны
    return nil
}

// Узел, который получает результат от NodeA
type NodeB struct {
    args   map[any]any
    result string
}

func (n *NodeB) DataType() any {
    return KeyDataB
}

func (n *NodeB) Dependencies() []any {
    return []any{KeyDataA}
}

func (n *NodeB) Execute(ctx context.Context) error {
    // Используем аргументы от NodeA
    if resultA, ok := n.args[KeyDataA].(string); ok {
        n.result = "processed: " + resultA
    }
    return nil
}

func (n *NodeB) GetResult() (any, error) {
    return n.result, nil
}

func (n *NodeB) SetArgs(args map[any]any) error {
    n.args = args
    return nil
}
```

**Важно:** 
- Узел с аргументами (`NodeWithArgs`) может зависеть только от узлов, которые также реализуют `NodeWithArgs` и возвращают результат
- Если узел с аргументами зависит от узла без `NodeWithArgs`, будет возвращена ошибка `ErrInvalidDependency`
- Узлы без аргументов могут зависеть от любых узлов (с аргументами или без)

## Параллельное выполнение

Метод `ExecuteInParallel` группирует узлы по рангам и выполняет узлы одного ранга параллельно:

- **Ранг 0:** узлы без зависимостей
- **Ранг 1:** узлы, зависящие только от узлов ранга 0
- **Ранг N:** узлы, зависящие только от узлов рангов 0..N-1

Узлы одного ранга выполняются параллельно, что может значительно ускорить выполнение для больших графов.

```go
// Выполнение с параллелизмом
err := graph.ExecuteInParallel(ctx)
```

Передача аргументов работает корректно и в параллельном режиме благодаря синхронизации доступа к результатам.

## Обработка ошибок

Пакет возвращает следующие ошибки:

- `ErrCircularDependency` - обнаружена циклическая зависимость между узлами
- `ErrFailedToSortNodes` - не удалось отсортировать узлы (обычно из-за циклических зависимостей)
- `ErrFailedToExecuteNode` - ошибка при выполнении операции узла
- `ErrInvalidDependency` - невалидная зависимость: узел с аргументами зависит от узла без результата

## Алгоритм

Используется алгоритм топологической сортировки Канна (Kahn's algorithm) для определения правильного порядка выполнения узлов с учетом их зависимостей.

