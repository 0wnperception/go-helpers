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
	tmpQueue        *queue.Queue[IDT]
	done            chan error
	ready           chan error
	err             chan error
}

func NewOperationManager[IDT comparable, TA any](maxops int) *OperationManager[IDT, TA] {
	return &OperationManager[IDT, TA]{
		locker:          &sync.RWMutex{},
		operations:      make(map[IDT]func(ctx context.Context, args TA) error, maxops),
		operationsQueue: queue.NewQueue[IDT](maxops),
		done:            make(chan error, 1),
		ready:           make(chan error, 1),
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
	m.tmpQueue = m.operationsQueue.Copy()
	for {
		select {
		case e := <-m.err:
			if e != nil {
				m.SetDone(e)
				m.setReady(nil)
				return e
			} else {
				if opID, ok := m.tmpQueue.Pull(); ok {
					go m.runOperation(ctx, opID, args, m.err)
				} else {
					if len(m.done) == 0 {
						m.SetDone(nil)
					}
					m.setReady(nil)
					return nil
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (m *OperationManager[IDT, TA]) SetDone(err error) {
	if len(m.done) > 0 {
		<-m.done
	}
	m.done <- err
}

func (m *OperationManager[IDT, TA]) setReady(err error) {
	if len(m.ready) > 0 {
		<-m.ready
	}
	m.ready <- err
}

func (m *OperationManager[IDT, TA]) GetProgressChans() (done <-chan error, ready <-chan error) {
	return m.done, m.ready
}

func (m *OperationManager[IDT, TA]) runOperation(ctx context.Context, id IDT, args TA, err chan error) {
	err <- m.operations[id](ctx, args)
}
