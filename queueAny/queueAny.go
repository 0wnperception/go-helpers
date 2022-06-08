package queueAny

import "sync"

//this package implements generic queue

type Queue[T any] struct {
	sync.Locker
	mem  []queueNode[T]
	head int
	tail int
	cap  int
	len  int
}

type queueNode[T any] struct {
	next int
	val  T
}

func NewQueue[T any](maxlen int) *Queue[T] {
	q := &Queue[T]{
		mem:    make([]queueNode[T], maxlen),
		Locker: &sync.RWMutex{},
		len:    0,
		cap:    maxlen,
	}
	for i := 0; i < maxlen-1; i++ {
		q.mem[i].next = i + 1
	}
	q.mem[maxlen-1].next = 0
	return q
}

func (q *Queue[T]) Copy() *Queue[T] {
	tmp := NewQueue[T](len(q.mem))
	copy(tmp.mem, q.mem)
	tmp.len = q.len
	tmp.head = q.head
	tmp.tail = q.tail
	return tmp
}

func (q *Queue[T]) Head() (t T, ok bool) {
	return q.mem[q.head].val, q.len > 0
}

func (q *Queue[T]) Push(v T) (ok bool) {
	if q.cap > 0 {
		q.Lock()
		if q.len == 0 {
			q.mem[q.head].val = v
			q.tail = q.head
		} else {
			q.tail = q.mem[q.tail].next
			q.mem[q.tail].val = v
		}
		q.len++
		q.cap--
		ok = true
		q.Unlock()
	}
	return
}

func (q *Queue[T]) Pull() (val T, ok bool) {
	if q.len > 0 {
		q.Lock()
		val = q.mem[q.head].val
		ok = true
		q.head = q.mem[q.head].next
		q.len--
		q.cap++
		q.Unlock()
	}
	return
}

func (q *Queue[T]) Len() int {
	return q.len
}

func (q *Queue[T]) List() []T {
	l := make([]T, q.len)
	for tmp, idx := q.head, 0; idx < q.len; tmp = q.mem[tmp].next {
		l[idx] = q.mem[tmp].val
		idx++
	}
	return l
}
