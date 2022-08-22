package repo

import (
	cache "github.com/0wnperception/go-helpers/pointer_cache"

	"github.com/pkg/errors"
)

type PersistentIface[T any, IDT comparable] interface {
	Store(ID IDT, e *T) error
	StoreTable(emap map[IDT]*T) error
	Delete(ID IDT) error
	GetList() ([]*T, error)
	IsExist(ID IDT) bool
	Flush()
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

func (repo *Repo[T, IDT]) Store(ID IDT, e *T) error {
	if repo.persistent != nil {
		if err := repo.persistent.Store(ID, e); err != nil {
			return errors.Wrap(err, "persistent error")
		}
	}
	repo.cache.Store(ID, e)
	return nil
}

func (repo *Repo[T, IDT]) StoreTable(emap map[IDT]*T) error {
	if repo.persistent != nil {
		if err := repo.persistent.StoreTable(emap); err != nil {
			return errors.Wrap(err, "persistent error")
		}
	}
	repo.cache.StoreTable(emap)
	return nil
}

func (repo *Repo[T, IDT]) StoreCache(ID IDT, e *T) {
	repo.cache.Store(ID, e)
}

func (repo *Repo[T, IDT]) Delete(ID IDT) error {
	if repo.persistent != nil {
		if err := repo.persistent.Delete(ID); err != nil {
			return err
		}
	}
	repo.DeleteCache(ID)
	return nil
}

func (repo *Repo[T, IDT]) DeleteCache(ID IDT) {
	repo.cache.Delete(ID)
}

func (repo *Repo[T, IDT]) GetByID(ID IDT) *T {
	return repo.cache.GetByID(ID)
}

func (repo *Repo[T, IDT]) GetCache() ([]*T, error) {
	return repo.cache.GetList(), nil
}

func (repo *Repo[T, IDT]) GetPersistent() ([]*T, error) {
	if repo.persistent != nil {
		if v, err := repo.persistent.GetList(); err != nil {
			return v, errors.Wrap(err, "persistent error")
		} else {
			return v, nil
		}
	} else {
		return nil, nil
	}
}

func (repo *Repo[T, IDT]) GetCacheLen() int {
	return repo.cache.Len()
}

func (repo *Repo[T, IDT]) IsExist(ID IDT) bool {
	return repo.cache.IsExist(ID)
}

func (repo *Repo[T, IDT]) Flush() {
	repo.persistent.Flush()
	repo.cache.Flush()
}
