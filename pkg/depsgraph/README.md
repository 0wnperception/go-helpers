# depsgraph

Пакет для управления графом зависимостей с топологической сортировкой. Позволяет выполнять операции над узлами в правильном порядке с учетом их зависимостей.

## Основные компоненты

- `Node` - интерфейс узла с зависимостями
- `Graph` - граф узлов с топологической сортировкой
- `AddNode` - добавление узла в граф
- `ExecuteAll` - выполнение операций всех узлов в правильном порядке

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

**Важно:** Узлы выполняются последовательно (не параллельно). Порядок между узлами с одинаковыми зависимостями (B и D) не гарантирован и может варьироваться между запусками.

## Обработка ошибок

Пакет возвращает следующие ошибки:

- `ErrCircularDependency` - обнаружена циклическая зависимость между узлами
- `ErrFailedToSortNodes` - не удалось отсортировать узлы (обычно из-за циклических зависимостей)
- `ErrFailedToExecuteNode` - ошибка при выполнении операции узла

## Алгоритм

Используется алгоритм топологической сортировки Канна (Kahn's algorithm) для определения правильного порядка выполнения узлов с учетом их зависимостей.

