package concurrent

import (
	"context"
	"sync/atomic"
)

type ConcurrentConfig struct {
	SimCapacity uint32
}

type Concurrent struct {
	busy     chan struct{}
	counter  uint32
	capacity uint32
}

var empty struct{} = struct{}{}

func NewConcurrent(cfg ConcurrentConfig) *Concurrent {
	return &Concurrent{
		busy:     make(chan struct{}, cfg.SimCapacity),
		capacity: cfg.SimCapacity,
	}
}

func (c *Concurrent) Borrow(ctx context.Context) (ok bool) {
	select {
	case c.busy <- empty:
		atomic.AddUint32(&c.counter, 1)
		ok = true
		break
	case <-ctx.Done():
		break
	}
	return
}

func (c *Concurrent) SettleUp() {
	if c.counter > 0 {
		atomic.AddUint32(&c.counter, ^uint32(0))
		<-c.busy
	}
}

func (c *Concurrent) IsAvailable() bool {
	return c.counter < c.capacity
}
