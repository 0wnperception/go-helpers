package priorityQueueAny

import "sync"

// this package implements generic priority queue

type PriorityQueueAnyIface[T any] interface {
	Push(priority int, v T) (ok bool)
	Pull() (t T, ok bool)
	Len() int
	Cap() int
	List() []T
	GetIterator() *pQueueAnyIterator
	Iterate(iter *pQueueAnyIterator) (t T, ok bool)
	PopByIterator(iter *pQueueAnyIterator) (old T, ok bool)
}

type pQueueAnyIterator struct {
	prev int
	idx  int
}

type pQueueAny[T any] struct {
	sync.Locker
	mem  []pQueueAnyNode[T]
	head int
	tail int
	cap  int
	len  int
	desc bool
}

type pQueueAnyNode[T any] struct {
	next     int
	val      T
	priority int
}

func NewPriorityQueueAny[T any](maxlen int, desc bool) PriorityQueueAnyIface[T] {
	q := &pQueueAny[T]{
		mem:    make([]pQueueAnyNode[T], maxlen),
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

func (q *pQueueAny[T]) Push(priority int, v T) (ok bool) {
	if q.cap > 0 {
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
				q.mem[tmp] = pQueueAnyNode[T]{
					val:      v,
					priority: priority,
					next:     free,
				}
			}
		}
		q.len++
		q.cap--
		ok = true
		q.Unlock()
	}
	return
}

func (q *pQueueAny[T]) Pull() (t T, ok bool) {
	if q.len > 0 {
		q.Lock()
		t = q.mem[q.head].val
		q.head = q.mem[q.head].next
		q.len--
		q.cap++
		ok = true
		q.Unlock()
	}
	return
}

func (q *pQueueAny[T]) Len() int {
	return q.len
}

func (q *pQueueAny[T]) Cap() int {
	return q.cap
}

func (q *pQueueAny[T]) List() []T {
	l := make([]T, q.len)
	for tmp, idx := q.head, 0; idx < q.len; tmp = q.mem[tmp].next {
		l[idx] = q.mem[tmp].val
		idx++
	}
	return l
}

func (q *pQueueAny[T]) GetIterator() *pQueueAnyIterator {
	return &pQueueAnyIterator{
		idx:  q.head,
		prev: q.head,
	}
}

func (q *pQueueAny[T]) Iterate(iter *pQueueAnyIterator) (t T, ok bool) {
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

func (q *pQueueAny[T]) PopByIterator(iter *pQueueAnyIterator) (old T, ok bool) {
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
