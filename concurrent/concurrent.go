package concurrent

import (
	"context"
	"sync/atomic"
)

type ConcurrentConfig struct {
	SimCapacity uint32
}

type Concurrent struct {
	busy    chan struct{}
	counter uint32
}

func NewConcurrent(cfg ConcurrentConfig) *Concurrent {
	return &Concurrent{
		busy: make(chan struct{}, cfg.SimCapacity),
	}
}

func (c *Concurrent) Borrow(ctx context.Context) {
	select {
	case c.busy <- struct{}{}:
		atomic.AddUint32(&c.counter, 1)
		break
	case <-ctx.Done():
		break
	}
}

func (c *Concurrent) SettleUp() {
	if c.counter > 0 {
		atomic.AddUint32(&c.counter, ^uint32(0))
		<-c.busy
	}
}
