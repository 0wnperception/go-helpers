package queueAny

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
	})
}
