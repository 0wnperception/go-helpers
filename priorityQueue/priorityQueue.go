package priorityQueue

import "sync"

// this package implements generic priority queue

type PriorityQueueIface[T comparable] interface {
	Push(priority int, v T) (ok bool)
	Pull() (t T, ok bool)
	Pop(v T) (old T, ok bool)
	Len() int
	Cap() int
	List() []T
	GetIterator() *pQueueIterator
	GetByIterator(iter *pQueueIterator) (t T, ok bool)
	Iterate(iter *pQueueIterator) (ok bool)
	PopByIterator(iter *pQueueIterator) (old T, ok bool)
}

type pQueueIterator struct {
	prev int
	idx  int
}

type pQueue[T comparable] struct {
	sync.Locker
	mem  []pQueueNode[T]
	head int
	tail int
	cap  int
	len  int
	desc bool
}

type pQueueNode[T comparable] struct {
	next     int
	val      T
	priority int
}

func NewPriorityQueue[T comparable](maxlen int, desc bool) PriorityQueueIface[T] {
	q := &pQueue[T]{
		mem:    make([]pQueueNode[T], maxlen),
		Locker: &sync.RWMutex{},
		len:    0,
		cap:    maxlen,
		desc:   desc,
	}
	for i := 0; i < maxlen-1; i++ {
		q.mem[i].next = i + 1
	}
	q.mem[maxlen-1].next = 0
	return q
}

func (q *pQueue[T]) Push(priority int, v T) (ok bool) {
	if q.cap > q.len {
		q.Lock()
		if q.len == 0 {
			q.mem[q.head].val = v
			q.mem[q.head].priority = priority
			q.tail = q.head
		} else {
			free := q.mem[q.tail].next
			tmp := q.head
			for idx := 0; idx < q.len; idx, tmp = idx+1, q.mem[tmp].next {
				if (priority < q.mem[tmp].priority && !q.desc) || (priority > q.mem[tmp].priority && q.desc) {
					break
				}
			}
			if tmp == free {
				q.mem[free].val = v
				q.mem[free].priority = priority
				q.tail = free
			} else {
				if tmp == q.tail {
					q.tail = free
					q.mem[free].val = q.mem[tmp].val
					q.mem[free].priority = q.mem[tmp].priority
				} else {
					q.mem[q.tail].next = q.mem[free].next
					q.mem[free] = q.mem[tmp]
				}
				q.mem[tmp] = pQueueNode[T]{
					val:      v,
					priority: priority,
					next:     free,
				}
			}
		}
		q.len++
		ok = true
		q.Unlock()
	}
	return
}

func (q *pQueue[T]) Pull() (t T, ok bool) {
	if q.len > 0 {
		q.Lock()
		t = q.mem[q.head].val
		q.head = q.mem[q.head].next
		q.len--
		ok = true
		q.Unlock()
	}
	return
}

func (q *pQueue[T]) Pop(v T) (old T, ok bool) {
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
				ok = true
				break
			}
		}
		q.Unlock()
	}
	return
}

func (q *pQueue[T]) Len() int {
	return q.len
}

func (q *pQueue[T]) Cap() int {
	return q.cap
}

func (q *pQueue[T]) List() []T {
	l := make([]T, q.len)
	for tmp, idx := q.head, 0; idx < q.len; tmp = q.mem[tmp].next {
		l[idx] = q.mem[tmp].val
		idx++
	}
	return l
}

func (q *pQueue[T]) GetIterator() *pQueueIterator {
	return &pQueueIterator{
		idx:  q.head,
		prev: -1,
	}
}

func (q *pQueue[T]) GetByIterator(iter *pQueueIterator) (t T, ok bool) {
	if iter != nil {
		if q.Len() > 0 {
			ok = iter.prev != q.tail
			if ok {
				t = q.mem[iter.idx].val
			}
		}
	}
	return
}

func (q *pQueue[T]) Iterate(iter *pQueueIterator) (ok bool) {
	if iter != nil {
		if q.Len() > 0 {
			ok = iter.prev != q.tail
			if ok {
				iter.prev = iter.idx
				iter.idx = q.mem[iter.idx].next
			}
		}
	}
	return
}

func (q *pQueue[T]) PopByIterator(iter *pQueueIterator) (old T, ok bool) {
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
		ok = true
		q.Unlock()
	}
	return
}
