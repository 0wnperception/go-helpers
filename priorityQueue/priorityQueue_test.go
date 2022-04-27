package priorityQueue

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

// BenchmarkPriorityQueue/push-8         	41448568	        28.18 ns/op	      96 B/op	       1 allocs/op
// BenchmarkPriorityQueue/pull-8         	566229372	         2.112 ns/op	       0 B/op	       0 allocs/op
func BenchmarkPriorityQueue(b *testing.B) {
	q := NewPriorityQueue[chan struct{}](b.N, false)
	b.Run("push", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			n := rand.Intn(b.N)
			q.Push(n, make(chan struct{}))
		}
	})
	b.Run("pull", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			q.Pull()
		}
	})
}

func TestPriorityQueue(t *testing.T) {
	r := require.New(t)

	t.Run("priority queue check asc order", func(t *testing.T) {
		q := NewPriorityQueue[string](10, false)
		var expected []string
		for i := 0; i < 10; i++ {
			v := rand.Intn(100)
			val := fmt.Sprintf("val %d", v)
			q.Push(v, val)
			expected = append(expected, val)
		}
		sort.Strings(expected)
		actual := q.List()
		r.Equal(expected, actual)
		t.Log("actual ", actual)
	})

	t.Run("priority queue check desc order", func(t *testing.T) {
		q := NewPriorityQueue[int](20, true)
		var expected []int
		for i := 0; i < 20; i++ {
			v := rand.Intn(200)
			q.Push(v, v)
			expected = append(expected, v)
		}
		sort.Slice(expected, func(i, j int) bool {
			return expected[i] > expected[j]
		})
		actual := q.List()
		r.Equal(expected, actual)
		t.Log("actual ", actual)
	})

	t.Run("priority queue check pull", func(t *testing.T) {
		q := NewPriorityQueue[int](10, false)
		var expected []int
		for i := 0; i < 10; i++ {
			v := rand.Intn(200)
			q.Push(v, v)
			expected = append(expected, v)
		}
		sort.Ints(expected)
		t.Log("initial ", q.List())
		v, _ := q.Pull()
		r.Equal(expected[0], v)
		actual := q.List()
		r.Equal(expected[1:], actual)
		t.Log("actual ", actual)
	})

	t.Run("priority queue check pop", func(t *testing.T) {
		q := NewPriorityQueue[int](10, false)
		var expected []int
		for i := 0; i < 10; i++ {
			v := rand.Intn(200)
			q.Push(v, v)
			expected = append(expected, v)
			r.Equal(i+1, q.Len())
		}
		sort.Ints(expected)
		t.Log("initial ", q.List())
		t.Log("pop ", expected[5])
		q.Pop(expected[5])
		actual := q.List()
		expected = append(expected[:5], expected[6:]...)
		r.Equal(expected, actual)
		t.Log("actual ", actual)
	})
}
