package operationManager

import (
	"context"
	"go-helpers/queue"
	"sync"
)

type OperationManager[IDT comparable, TA any] struct {
	locker          sync.Locker
	operations      map[IDT]func(ctx context.Context, args TA) error
	operationsQueue *queue.Queue[IDT]
}

func NewOperationManager[IDT comparable, TA any](maxops int) *OperationManager[IDT, TA] {
	return &OperationManager[IDT, TA]{
		locker:          &sync.RWMutex{},
		operations:      make(map[IDT]func(ctx context.Context, args TA) error, maxops),
		operationsQueue: queue.NewQueue[IDT](maxops),
	}
}

func (m *OperationManager[IDT, TA]) AddOperation(ID IDT, op func(ctx context.Context, args TA) error) {
	m.locker.Lock()
	m.operations[ID] = op
	m.locker.Unlock()
	m.operationsQueue.Push(ID)
}

func (m *OperationManager[IDT, TA]) Run(ctx context.Context, args TA) error {
	err := make(chan error, 1)
	err <- nil
	for {
		select {
		case e := <-err:
			if e != nil {
				return e
			} else {
				if opID, ok := m.operationsQueue.Pull(); ok {
					go m.runOperation(ctx, opID, args, err)
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (m *OperationManager[IDT, TA]) runOperation(ctx context.Context, id IDT, args TA, err chan error) {
	err <- m.operations[id](ctx, args)
}
