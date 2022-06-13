package operationManager

import (
	"context"
	"sync"

	"github.com/0wnperception/go-helpers/concurrent"
	"github.com/0wnperception/go-helpers/queue"
)

type OperationManager[IDT comparable, TP any] struct {
	locker          sync.Locker
	operations      map[IDT]func(ctx context.Context, params TP) error
	operationsQueue queue.QueueIface[IDT]
	tmpQueue        queue.QueueIface[IDT]
	done            chan error
	ready           chan error
	err             chan error
}

func NewOperationManager[IDT comparable, TP any](maxops int) *OperationManager[IDT, TP] {
	return &OperationManager[IDT, TP]{
		locker:          &sync.RWMutex{},
		operations:      make(map[IDT]func(ctx context.Context, params TP) error, maxops),
		operationsQueue: queue.NewQueue[IDT](maxops),
		done:            make(chan error, 1),
		ready:           make(chan error, 1),
		err:             make(chan error, 1),
	}
}

func (m *OperationManager[IDT, TP]) AddOperation(ID IDT, op func(ctx context.Context, params TP) error) {
	m.locker.Lock()
	m.operations[ID] = op
	m.locker.Unlock()
	m.operationsQueue.Push(ID)
}

func (m *OperationManager[IDT, TP]) Run(ctx context.Context, concurrent *concurrent.Concurrent, params TP) (done <-chan error, ready <-chan error) {
	if concurrent != nil {
		if ok := concurrent.Borrow(ctx); ok {
			go m.run(ctx, concurrent, params)
			return m.done, m.ready
		}
	} else {
		go m.run(ctx, concurrent, params)
		return m.done, m.ready
	}
	return nil, nil
}

func (m *OperationManager[IDT, TP]) run(ctx context.Context, concurrent *concurrent.Concurrent, params TP) {
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
				m.setReady(concurrent, nil)
				return
			} else {
				if opID, ok := m.tmpQueue.Pull(); ok {
					go m.runOperation(ctx, opID, params, m.err)
				} else {
					if len(m.done) == 0 {
						m.SetDone(nil)
					}
					m.setReady(concurrent, nil)
					return
				}
			}
		case <-ctx.Done():
			m.setReady(concurrent, nil)
			return
		}
	}
}

func (m *OperationManager[IDT, TP]) SetDone(err error) {
	if len(m.done) > 0 {
		<-m.done
	}
	m.done <- err
}

func (m *OperationManager[IDT, TP]) setReady(concurrent *concurrent.Concurrent, err error) {
	if len(m.ready) > 0 {
		<-m.ready
	}
	m.ready <- err
	if concurrent != nil {
		concurrent.SettleUp()
	}
}

func (m *OperationManager[IDT, TP]) runOperation(ctx context.Context, id IDT, params TP, err chan error) {
	err <- m.operations[id](ctx, params)
}
