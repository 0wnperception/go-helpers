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
	err             chan error
}

func NewOperationManager[IDT comparable, TA any](maxops int) *OperationManager[IDT, TA] {
	return &OperationManager[IDT, TA]{
		locker:          &sync.RWMutex{},
		operations:      make(map[IDT]func(ctx context.Context, args TA) error, maxops),
		operationsQueue: queue.NewQueue[IDT](maxops),
		err:             make(chan error, 1),
	}
}

func (m *OperationManager[IDT, TA]) AddOperation(ID IDT, op func(ctx context.Context, args TA) error) {
	m.locker.Lock()
	m.operations[ID] = op
	m.locker.Unlock()
	m.operationsQueue.Push(ID)
}

func (m *OperationManager[IDT, TA]) Run(ctx context.Context, args TA) error {
	if len(m.err) > 0 {
		<-m.err
	}
	m.err <- nil
	for {
		select {
		case e := <-m.err:
			if e != nil {
				return e
			} else {
				if opID, ok := m.operationsQueue.Pull(); ok {
					go m.runOperation(ctx, opID, args, m.err)
				} else {
					return nil
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
