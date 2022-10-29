package pointers_storage

import (
	"sync"
)

type PointersStorageIface[T any, IDT comparable] interface {
	Store(ID IDT, e *T)
	StoreTable(emap map[IDT]*T)
	Delete(ID IDT)
	GetByID(ID IDT) *T
	GetList() []*T
	IsExist(ID IDT) bool
	Len() int
	Flush()
}

type pStorage[T any, IDT comparable] struct {
	sync.Locker
	Map map[IDT]*T
}

func NewPointersStorage[T any, IDT comparable]() *pStorage[T, IDT] {
	return &pStorage[T, IDT]{
		Locker: &sync.Mutex{},
		Map:    make(map[IDT]*T),
	}
}

func (c *pStorage[T, IDT]) Store(ID IDT, e *T) {
	c.Lock()
	c.Map[ID] = e
	c.Unlock()
}

func (c *pStorage[T, IDT]) StoreTable(emap map[IDT]*T) {
	c.Lock()
	for key, val := range emap {
		c.Map[key] = val
	}
	c.Unlock()
}

func (c *pStorage[T, IDT]) Delete(ID IDT) {
	c.Lock()
	delete(c.Map, ID)
	c.Unlock()
}

func (c *pStorage[T, IDT]) Flush() {
	c.Map = make(map[IDT]*T, len(c.Map))
}

func (c *pStorage[T, IDT]) GetByID(ID IDT) *T {
	if e, ok := c.Map[ID]; ok {
		return e
	}
	return nil
}

func (c *pStorage[T, IDT]) GetList() (l []*T) {
	l = make([]*T, len(c.Map))
	idx := 0
	for _, e := range c.Map {
		l[idx] = e
		idx++
	}
	return l
}

func (c *pStorage[T, IDT]) IsExist(ID IDT) bool {
	_, ok := c.Map[ID]
	return ok
}

func (c *pStorage[T, IDT]) Len() int {
	return len(c.Map)
}
