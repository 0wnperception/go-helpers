package queue

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

// BenchmarkQueue/push-8         	55879144	        21.24 ns/op	      96 B/op	       1 allocs/op
// BenchmarkQueue/pull-8         	568340923	         2.117 ns/op	       0 B/op	       0 allocs/op
func BenchmarkQueue(b *testing.B) {
	q := NewQueue[chan struct{}](b.N)
	b.Run("push", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			q.Push(make(chan struct{}))
		}
	})
	b.Run("pull", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			q.Pull()
		}
	})
}

// BenchmarkChanPush/push-8         	37335199	        42.22 ns/op	      96 B/op	       1 allocs/op
// BenchmarkChanPush/pull-8         	84343450	        13.75 ns/op	       0 B/op	       0 allocs/op
func BenchmarkChan(b *testing.B) {
	maxlen := 100000000
	q := make(chan chan struct{}, maxlen)
	b.Run("push", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			q <- make(chan struct{})
		}
	})
	for i := len(q); i < maxlen; i++ {
		q <- make(chan struct{})
	}
	b.Run("pull", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			<-q
		}
	})
}

func TestQueue(t *testing.T) {
	r := require.New(t)

	t.Run("queue check push", func(t *testing.T) {
		q := NewQueue[string](10)
		r.Equal(10, q.Cap())
		var expected []string
		for i := 0; i < 10; i++ {
			v := rand.Intn(100)
			val := fmt.Sprintf("val %d", v)
			q.Push(val)
			expected = append(expected, val)
		}
		actual := q.List()
		r.Equal(expected, actual)
		t.Log("actual ", actual)
	})

	t.Run("queue check pull", func(t *testing.T) {
		q := NewQueue[int](10)
		var expected []int
		for i := 0; i < 10; i++ {
			v := rand.Intn(200)
			q.Push(v)
			expected = append(expected, v)
		}
		t.Log("initial ", q.List())
		v, _ := q.Pull()
		r.Equal(expected[0], v)
		actual := q.List()
		r.Equal(expected[1:], actual)
		t.Log("actual ", actual)
		r.Equal(10, q.Cap())
		r.Equal(9, q.Len())
	})

	t.Run("queue check pop", func(t *testing.T) {
		q := NewQueue[int](10)
		var expected []int
		for i := 0; i < 10; i++ {
			v := rand.Intn(200)
			q.Push(v)
			expected = append(expected, v)
			r.Equal(i+1, q.Len())
		}
		t.Log("initial ", q.List())
		t.Log("pop ", expected[5])
		q.Pop(expected[5])
		actual := q.List()
		expected = append(expected[:5], expected[6:]...)
		r.Equal(expected, actual)
		t.Log("actual ", actual)
	})

	t.Run("queue check iterator", func(t *testing.T) {
		t.Run("queue check good iterator ones", func(t *testing.T) {
			q := NewQueue[string](1)
			var expected []string
			for i := 0; i < 1; i++ {
				v := rand.Intn(100)
				val := fmt.Sprintf("val %d", v)
				q.Push(val)
				expected = append(expected, val)
			}
			iter := q.GetIterator()
			actual := []string{}
			for i := 0; i < q.Len(); i++ {
				v, ok := q.GetByIterator(iter)
				if ok {
					actual = append(actual, v)
				} else {
					break
				}
				q.Iterate(iter)
			}
			r.Equal(expected, actual)
		})

		t.Run("queue check good iterator on array", func(t *testing.T) {
			q := NewQueue[string](10)
			var expected []string
			for i := 0; i < 10; i++ {
				v := rand.Intn(100)
				val := fmt.Sprintf("val %d", v)
				q.Push(val)
				expected = append(expected, val)
			}
			_, ok := q.Pull()
			r.True(ok)
			_, ok = q.Pull()
			r.True(ok)
			expected = expected[2:]
			iter := q.GetIterator()
			actual := []string{}
			for i := 0; i < q.Len(); i++ {
				v, ok := q.GetByIterator(iter)
				if ok {
					actual = append(actual, v)
				} else {
					break
				}
				q.Iterate(iter)
			}
			r.Equal(expected, actual)
		})

		t.Run("queue check iterator empty", func(t *testing.T) {
			q := NewQueue[string](10)
			iter := q.GetIterator()
			ok := q.Iterate(iter)
			r.False(ok)
		})

		t.Run("queue check nil iterator", func(t *testing.T) {
			q := NewQueue[string](10)
			ok := q.Iterate(nil)
			r.False(ok)
		})

		t.Run("queue check pop by iterator", func(t *testing.T) {
			q := NewQueue[int](10)
			var expected []int
			popElem := 0
			for i := 0; i < 10; i++ {
				v := rand.Intn(200)
				q.Push(v)
				expected = append(expected, v)
				r.Equal(i+1, q.Len())
				if i == 6 {
					popElem = v
				}
			}

			iter := q.GetIterator()
			for i := 0; i < q.Len(); i++ {
				v, ok := q.GetByIterator(iter)
				t.Log("iterate ", v, ok)
				if ok && v == popElem {
					break
				}
				q.Iterate(iter)
			}
			t.Log("pop ", popElem)
			t.Log("initial ", q.List())
			q.PopByIterator(iter)
			actual := q.List()
			expected = append(expected[:6], expected[7:]...)
			r.Equal(expected, actual)
			t.Log("actual ", actual)
		})
	})

}
