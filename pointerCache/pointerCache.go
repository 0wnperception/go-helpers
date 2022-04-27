package cache

import (
	"sync"
)

type pCache[T any, IDT comparable] struct {
	sync.Locker
	Map map[IDT]*T
}

func NewPointerCache[T any, IDT comparable]() *pCache[T, IDT] {
	return &pCache[T, IDT]{
		Locker: &sync.RWMutex{},
		Map:    make(map[IDT]*T),
	}
}

func (c *pCache[T, IDT]) Store(ID IDT, e *T) {
	c.Lock()
	c.Map[ID] = e
	c.Unlock()
}

func (c *pCache[T, IDT]) StoreTable(emap map[IDT]*T) {
	for key, val := range emap {
		c.Store(key, val)
	}
}

func (c *pCache[T, IDT]) Delete(ID IDT) {
	c.Lock()
	delete(c.Map, ID)
	c.Unlock()
}

func (c *pCache[T, IDT]) Flush() {
	c.Map = make(map[IDT]*T, len(c.Map))
}

func (c *pCache[T, IDT]) GetByID(ID IDT) *T {
	if e, ok := c.Map[ID]; ok {
		return e
	}
	return nil
}

func (c *pCache[T, IDT]) IsExist(ID IDT) bool {
	_, ok := c.Map[ID]
	return ok
}

func (c *pCache[T, IDT]) Len() int {
	return len(c.Map)
}
