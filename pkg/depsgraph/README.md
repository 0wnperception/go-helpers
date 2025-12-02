# depsgraph

Пакет для управления графом зависимостей с топологической сортировкой. Позволяет выполнять операции над узлами в правильном порядке с учетом их зависимостей.

## Основные компоненты

- `Node[T]` - интерфейс узла с зависимостями
- `Graph` - граф узлов с топологической сортировкой
- `AddNode[T]` - добавление узла в граф
- `ExecuteAll` - выполнение операций всех узлов в правильном порядке

## Пример использования

```go
package main

import (
    "context"
    "reflect"
    
    "github.com/0wnperception/go-helpers/pkg/depsgraph"
)

// Определяем типы данных для узлов
type DataA struct{}
type DataB struct{}
type DataC struct{}
type DataD struct{}

// Реализуем узел без зависимостей
type NodeA struct{}

func (n *NodeA) DataType() reflect.Type {
    return reflect.TypeOf((*DataA)(nil)).Elem()
}

func (n *NodeA) Dependencies() []reflect.Type {
    return nil // нет зависимостей
}

func (n *NodeA) Execute(ctx context.Context) error {
    // выполнение операций для узла A
    return nil
}

// Реализуем узел с одной зависимостью
type NodeB struct{}

func (n *NodeB) DataType() reflect.Type {
    return reflect.TypeOf((*DataB)(nil)).Elem()
}

func (n *NodeB) Dependencies() []reflect.Type {
    return []reflect.Type{
        reflect.TypeOf((*DataA)(nil)).Elem(), // зависит от DataA
    }
}

func (n *NodeB) Sync(ctx context.Context) error {
    // выполнение операций для узла B
    return nil
}

// Реализуем узел с несколькими зависимостями
type NodeC struct{}

func (n *NodeC) DataType() reflect.Type {
    return reflect.TypeOf((*DataC)(nil)).Elem()
}

func (n *NodeC) Dependencies() []reflect.Type {
    return []reflect.Type{
        reflect.TypeOf((*DataA)(nil)).Elem(), // зависит от DataA
        reflect.TypeOf((*DataB)(nil)).Elem(), // зависит от DataB
    }
}

func (n *NodeC) Sync(ctx context.Context) error {
    // выполнение операций для узла C
    return nil
}

// Реализуем независимый узел
type NodeD struct{}

func (n *NodeD) DataType() reflect.Type {
    return reflect.TypeOf((*DataD)(nil)).Elem()
}

func (n *NodeD) Dependencies() []reflect.Type {
    return []reflect.Type{
        reflect.TypeOf((*DataA)(nil)).Elem(), // зависит от DataA
    }
}

func (n *NodeD) Sync(ctx context.Context) error {
    // выполнение операций для узла D
    return nil
}

func ProcessAll(ctx context.Context) error {
    // Создаем граф
    graph := depsgraph.NewGraph()

    // Добавляем узлы в произвольном порядке
    depsgraph.AddNode[DataA](graph, &NodeA{})
    depsgraph.AddNode[DataB](graph, &NodeB{})
    depsgraph.AddNode[DataC](graph, &NodeC{})
    depsgraph.AddNode[DataD](graph, &NodeD{})

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

