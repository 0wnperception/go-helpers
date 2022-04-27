package workerPoll

import (
	"context"
	"errors"
	"robot_agent/pkg/priorityQueue"
	"robot_agent/pkg/queue"
	"sync"
	"time"
)

type WorkerPool[T any, IDT comparable] struct {
	sync.Locker
	m map[IDT]*workerNode[T]
	q *priorityQueue.PQueue[IDT]
}

type workerNode[T any] struct {
	busy       bool
	requestedQ *queue.Queue[chan struct{}]
	v          *T
}

func NewWorkerPool[T any, IDT comparable](desc bool) *WorkerPool[T, IDT] {
	return &WorkerPool[T, IDT]{
		Locker: &sync.RWMutex{},
		m:      make(map[IDT]*workerNode[T]),
		q:      priorityQueue.NewPriorityQueue[IDT](10, desc),
	}
}

func (p *WorkerPool[T, IDT]) Register(id IDT, e *T) error {
	if e == nil {
		return errors.New("invalid element")
	}
	if _, ok := p.m[id]; ok {
		return errors.New("id already exist")
	}
	p.Lock()
	p.m[id] = &workerNode[T]{
		requestedQ: queue.NewQueue[chan struct{}](8),
		v:          e,
	}
	p.Unlock()
	return nil
}

func (p *WorkerPool[T, IDT]) Unregister(id IDT) error {
	p.RemoveFromQueue(id)
	p.Lock()
	delete(p.m, id)
	p.Unlock()
	return nil
}

func (p *WorkerPool[T, IDT]) GetFromQueue() (val *T) {
	if id, ok := p.q.Pull(); ok {
		val = p.m[id].v
	}
	return
}

func (p *WorkerPool[T, IDT]) RequestByID(ctx context.Context, id IDT) *T {
	if node, ok := p.m[id]; ok {
		if !node.busy {
			node.busy = true
			p.RemoveFromQueue(id)
			return node.v
		} else {
			var ok bool = false
			ch := make(chan struct{})
			t := time.NewTicker(20 * time.Second)
			for !ok {
				ok := node.requestedQ.Push(ch)
				if !ok {
					select {
					case <-t.C:
						break
					case <-ctx.Done():
						return nil
					}
				}
			}
			select {
			case <-ch:
				return node.v
			case <-ctx.Done():
				return nil
			}
		}
	} else {
		return nil
	}
}

func (p *WorkerPool[T, IDT]) ReturnToQueue(priority int, id IDT) (err error) {
	if node, ok := p.m[id]; ok {
		if node.busy {
			if node.requestedQ.Len() > 0 {
				ch, _ := node.requestedQ.Pull()
				close(ch)
			} else {
				node.busy = false
				p.q.Push(priority, id)
			}
		} else {
			err = errors.New("worker already free")
		}
	} else {
		err = errors.New("not registered element")
	}
	return
}

func (p *WorkerPool[T, IDT]) RemoveFromQueue(id IDT) {
	p.q.Pop(id)
}
