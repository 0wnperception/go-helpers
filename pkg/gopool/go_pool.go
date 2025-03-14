package gopool

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"time"

	"github.com/panjf2000/ants"
)

const noexpiry = 72 * time.Hour

type GoPool struct {
	wg      *sync.WaitGroup
	errSync *sync.Mutex
	err     error
	pool    *ants.Pool
}

func New() (*GoPool, error) {
	pool, err := ants.NewPool(runtime.NumCPU(), ants.WithExpiryDuration(noexpiry))
	if err != nil {
		return nil, err
	}

	return &GoPool{
		wg:      &sync.WaitGroup{},
		errSync: &sync.Mutex{},
		pool:    pool,
	}, nil
}

func (p *GoPool) Run(ctx context.Context, f func(ctx context.Context) error) error {
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

func (p *GoPool) Wait() error {
	p.wg.Wait()
	p.pool.Release()

	return p.err
}
