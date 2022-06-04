package priorityQueue

import "sync"

//this package implements generic priority queue

type PQueue[T comparable] struct {
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

func NewPriorityQueue[T comparable](maxlen int, desc bool) *PQueue[T] {
	q := &PQueue[T]{
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

func (q *PQueue[T]) Push(priority int, v T) (ok bool) {
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
				q.mem[tmp] = pQueueNode[T]{
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

func (q *PQueue[T]) Pull() (t T, ok bool) {
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

func (q *PQueue[T]) Pop(v T) (old T, ok bool) {
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

func (q *PQueue[T]) Len() int {
	return q.len
}

func (q *PQueue[T]) List() []T {
	l := make([]T, q.len)
	for tmp, idx := q.head, 0; idx < q.len; tmp = q.mem[tmp].next {
		l[idx] = q.mem[tmp].val
		idx++
	}
	return l
}
