package concurrent

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/0wnperception/go-helpers/pkg/queue"
)

type ConcurrentConfig struct {
	SimCapacity uint32
}

type Concurrent struct {
	locker   sync.Locker
	queue    queue.QueueIface[chan struct{}]
	counter  uint32
	users    uint32
	capacity uint32
}

func NewConcurrent(cfg ConcurrentConfig, maxCap int) *Concurrent {
	return &Concurrent{
		locker:   &sync.RWMutex{},
		queue:    queue.NewQueue[chan struct{}](maxCap),
		capacity: cfg.SimCapacity,
	}
}

func (c *Concurrent) Borrow(ctx context.Context) (ok bool) {
	if c.IsAvailable() {
		atomic.AddUint32(&c.users, 1)
		return true
	} else {
		c.locker.Lock()
		ch := make(chan struct{})
		c.queue.Push(ch)
		c.locker.Unlock()
		select {
		case <-ch:
			return true
		case <-ctx.Done():
			c.queue.Pop(ch)
		}
	}
	return
}

func (c *Concurrent) SettleUp() {
	if c.users > 0 {
		c.locker.Lock()
		if ch, ok := c.queue.Pull(); ok {
			close(ch)
		} else {
			atomic.AddUint32(&c.users, ^uint32(0))
		}
		c.locker.Unlock()
	}
}

func (c *Concurrent) IsAvailable() bool {
	return c.users < c.capacity
}
