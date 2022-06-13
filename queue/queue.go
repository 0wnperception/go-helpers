package queue

import "sync"

// this package implements generic queue

type QueueIface[T comparable] interface {
	Head() (t T, ok bool)
	Push(v T) (ok bool)
	Pull() (val T, ok bool)
	Pop(v T) (old T, ok bool)
	Copy() QueueIface[T]
	Len() int
	Cap() int
	List() []T
	GetIterator() *queueIterator
	Iterate(iter *queueIterator) (t T, ok bool)
	PopByIterator(iter *queueIterator) (old T, ok bool)
}

type queueIterator struct {
	prev int
	idx  int
}

type queue[T comparable] struct {
	sync.Locker
	mem  []queueNode[T]
	head int
	tail int
	cap  int
	len  int
}

type queueNode[T comparable] struct {
	next int
	val  T
}

func NewQueue[T comparable](maxlen int) QueueIface[T] {
	q := &queue[T]{
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

func (q *queue[T]) Copy() QueueIface[T] {
	tmp := &queue[T]{
		mem:    make([]queueNode[T], len(q.mem)),
		Locker: &sync.RWMutex{},
		len:    0,
		cap:    len(q.mem),
	}
	copy(tmp.mem, q.mem)
	tmp.len = q.len
	tmp.head = q.head
	tmp.tail = q.tail
	return tmp
}

func (q *queue[T]) Head() (t T, ok bool) {
	return q.mem[q.head].val, q.len > 0
}

func (q *queue[T]) Push(v T) (ok bool) {
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

func (q *queue[T]) Pull() (val T, ok bool) {
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

func (q *queue[T]) Pop(v T) (old T, ok bool) {
	if q.len > 0 {
		q.Lock()
		for idx, tmp, prev := 0, q.head, q.head; idx < q.len; idx, prev, tmp = idx+1, tmp, q.mem[tmp].next {
			if q.mem[tmp].val == v {
				old = v
				switch tmp {
				case q.head:
					q.head = q.mem[q.head].next
					break
				case q.tail:
					q.tail = prev
					break
				default:
					q.mem[prev].next = q.mem[tmp].next
					q.mem[tmp].next = q.mem[q.tail].next
					q.mem[q.tail].next = tmp
				}
				q.len--
				q.cap++
				ok = true
				break
			}
		}
		q.Unlock()
	}
	return
}

func (q *queue[T]) Len() int {
	return q.len
}

func (q *queue[T]) Cap() int {
	return q.cap
}

func (q *queue[T]) List() []T {
	l := make([]T, q.len)
	for tmp, idx := q.head, 0; idx < q.len; tmp = q.mem[tmp].next {
		l[idx] = q.mem[tmp].val
		idx++
	}
	return l
}

func (q *queue[T]) GetIterator() *queueIterator {
	return &queueIterator{
		idx:  q.head,
		prev: q.head,
	}
}

func (q *queue[T]) Iterate(iter *queueIterator) (t T, ok bool) {
	if iter != nil {
		if iter.idx >= 0 && iter.idx < q.Len() && iter.prev >= 0 && iter.prev < q.Len() {
			ok = iter.prev != q.tail
			if ok {
				t = q.mem[iter.idx].val
				iter.prev = iter.idx
				iter.idx = q.mem[iter.idx].next
			}
		}
	}
	return
}

func (q *queue[T]) PopByIterator(iter *queueIterator) (old T, ok bool) {
	if q.len > 0 {
		q.Lock()
		old = q.mem[iter.idx].val
		switch iter.idx {
		case q.head:
			q.head = q.mem[q.head].next
			break
		case q.tail:
			q.tail = iter.prev
			break
		default:
			q.mem[iter.prev].next = q.mem[iter.idx].next
			q.mem[iter.idx].next = q.mem[q.tail].next
			q.mem[q.tail].next = iter.idx
		}
		q.len--
		q.cap++
		ok = true
		q.Unlock()
	}
	return
}
