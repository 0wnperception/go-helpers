package repo

import (
	"context"
	cache "go-helpers/pointerCache"

	"github.com/pkg/errors"
)

type PersistentIface[T any, IDT comparable] interface {
	Store(ctx context.Context, ID IDT, e *T) error
	StoreTable(ctx context.Context, emap map[IDT]*T) error
	Delete(ctx context.Context, ID IDT) error
	GetList(ctx context.Context) ([]*T, error)
	IsExist(ctx context.Context, ID IDT) bool
	Flush(ctx context.Context)
}

type PointerCacheIface[T any, IDT comparable] interface {
	Store(ID IDT, e *T)
	StoreTable(emap map[IDT]*T)
	Delete(ID IDT)
	GetByID(ID IDT) *T
	GetList() []*T
	IsExist(ID IDT) bool
	Len() int
	Flush()
}

type Repo[T any, IDT comparable] struct {
	cache      PointerCacheIface[T, IDT]
	persistent PersistentIface[T, IDT]
}

func NewRepo[T any, IDT comparable](pers PersistentIface[T, IDT]) *Repo[T, IDT] {
	return &Repo[T, IDT]{
		cache:      cache.NewPointerCache[T, IDT](),
		persistent: pers,
	}
}

func (repo *Repo[T, IDT]) Store(ctx context.Context, ID IDT, e *T) error {
	if repo.persistent != nil {
		if err := repo.persistent.Store(ctx, ID, e); err != nil {
			return errors.Wrap(err, "persistent error")
		}
	}
	repo.cache.Store(ID, e)
	return nil
}

func (repo *Repo[T, IDT]) StoreTable(ctx context.Context, emap map[IDT]*T) error {
	if repo.persistent != nil {
		if err := repo.persistent.StoreTable(ctx, emap); err != nil {
			return errors.Wrap(err, "persistent error")
		}
	}
	repo.cache.StoreTable(emap)
	return nil
}

func (repo *Repo[T, IDT]) StoreCache(ctx context.Context, ID IDT, e *T) {
	repo.cache.Store(ID, e)
}

func (repo *Repo[T, IDT]) Delete(ctx context.Context, ID IDT) error {
	if repo.persistent != nil {
		if err := repo.persistent.Delete(ctx, ID); err != nil {
			return err
		}
	}
	repo.DeleteCache(ctx, ID)
	return nil
}

func (repo *Repo[T, IDT]) DeleteCache(ctx context.Context, ID IDT) {
	repo.cache.Delete(ID)
}

func (repo *Repo[T, IDT]) GetByID(ctx context.Context, ID IDT) *T {
	return repo.cache.GetByID(ID)
}

func (repo *Repo[T, IDT]) GetCache(ctx context.Context) ([]*T, error) {
	return repo.cache.GetList(), nil
}

func (repo *Repo[T, IDT]) GetPersistent(ctx context.Context) ([]*T, error) {
	if repo.persistent != nil {
		if v, err := repo.persistent.GetList(ctx); err != nil {
			return v, errors.Wrap(err, "persistent error")
		} else {
			return v, nil
		}
	} else {
		return nil, nil
	}
}

func (repo *Repo[T, IDT]) GetCacheLen(ctx context.Context) int {
	return repo.cache.Len()
}

func (repo *Repo[T, IDT]) IsExist(ctx context.Context, ID IDT) bool {
	return repo.cache.IsExist(ID)
}

func (repo *Repo[T, IDT]) Flush(ctx context.Context) {
	repo.persistent.Flush(ctx)
	repo.cache.Flush()
}
