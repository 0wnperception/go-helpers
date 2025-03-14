package gopool

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants"
)

const noexpiry = 72 * time.Hour

var (
	ErrNotStarted      = errors.New("not started")
	ErrAlreadyStarted  = errors.New("already started")
	ErrAlreadyFinished = errors.New("already finished")
)

type GoPool struct {
	wg                *sync.WaitGroup
	errSync           *sync.Mutex
	err               error
	pool              *ants.Pool
	cancel            func()
	cancelOnce        *sync.Once
	started, finished atomic.Bool
}

func New() (*GoPool, error) {
	pool, err := ants.NewPool(runtime.NumCPU(), ants.WithExpiryDuration(noexpiry))
	if err != nil {
		return nil, err
	}

	return &GoPool{
		wg:         &sync.WaitGroup{},
		errSync:    &sync.Mutex{},
		pool:       pool,
		cancelOnce: &sync.Once{},
	}, nil
}

func (p *GoPool) Run(ctx context.Context, f func(ctx context.Context) error) error {
	if p.finished.Load() {
		return ErrAlreadyFinished
	}

	if p.started.Load() {
		return ErrAlreadyStarted
	}

	p.started.Store(true)

	ctx, p.cancel = context.WithCancel(ctx)
	p.wg.Add(1)

	if err := p.pool.Submit(func() {
		if err := f(ctx); err != nil {
			p.errSync.Lock()
			p.err = errors.Join(p.err, err)
			p.errSync.Unlock()
		}

		p.wg.Done()
	}); err != nil {
		p.errSync.Lock()
		p.err = errors.Join(p.err, err)
		p.errSync.Unlock()

		p.wg.Done()

		return err
	}

	return nil
}

func (p *GoPool) Cancel() {
	if p.started.Load() {
		p.cancelOnce.Do(func() {
			p.cancel()
			p.cancel = nil
		})
	}
}

func (p *GoPool) WaitFinish() error {
	if p.finished.Load() {
		return ErrAlreadyFinished
	}

	if !p.started.Load() {
		return ErrNotStarted
	}

	p.wg.Wait()
	p.pool.Release()
	p.finished.Store(true)

	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}

	return p.err
}
