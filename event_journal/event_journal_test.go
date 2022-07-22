package event_journal

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSimple(t *testing.T) {
	r := require.New(t)

	t.Run("put", func(t *testing.T) {
		j := NewEventJournal(3, nil)
		r.Equal(1, int(j.BasketsCounter))
		r.Equal(3, int(j.InMemoryBasket.Size))
		r.Equal(1, int(j.InMemoryBasket.Number))
		r.Len(j.BasketsIDsList, 1)
		r.Len(j.InMemoryBasket.Events, 3)

		j.PutEvent("some_event/1", []byte{})
		r.Equal(1, int(j.InMemoryBasket.Number))
		r.Equal(1, int(j.EventsCounter))
		r.Len(j.BasketsIDsList, 1)
		r.Len(j.InMemoryBasket.Events, 3)
		t.Log(j.InMemoryBasket.Events)

		j.PutEvent("some_event/2", []byte{})
		j.PutEvent("some_event/3", []byte{})
		r.Equal(3, int(j.EventsCounter))
		r.Equal(1, int(j.InMemoryBasket.Number))
		t.Log(j.InMemoryBasket.Events)

		j.PutEvent("some_event/4", []byte{})
		r.Equal(2, int(j.InMemoryBasket.Number))
		r.Equal(2, int(j.BasketsCounter))
		t.Log(j.InMemoryBasket.Events)
	})

	t.Run("get", func(t *testing.T) {
		j := NewEventJournal(3, nil)
		j.PutEvent("some_event/1", []byte{123})
		c := j.GetEventCounter()
		e, err := j.GetEventByCounter(c)
		r.NoError(err)
		t.Log(e)
		t.Log(j.InMemoryBasket.Events)
	})

	t.Run("wait", func(t *testing.T) {
		j := NewEventJournal(4, nil)
		j.PutEvent("some_event/1", []byte{123})
		c := j.GetEventCounter()
		go func() {
			for i := 0; i < 3; i++ {
				time.Sleep(time.Millisecond)
				t.Log(fmt.Sprintf("put some_event/%d", i+1))
				j.PutEvent(fmt.Sprintf("some_event/%d", i+1), []byte{byte(i)})
			}
		}()
		var err error
		for i := 0; i < 3; i++ {
			c, err = j.WaitEvent(c, "some_event/", context.Background())
			r.NoError(err)
			e, err := j.GetEventByCounter(c)
			r.NoError(err)
			r.Equal([]byte{byte(i)}, e.Payload)
			t.Log(e)
		}
	})

	t.Run("parallel", func(t *testing.T) {
		n := 20
		uNum := 5
		j := NewEventJournal(n, nil)
		c := j.GetEventCounter()
		for i := 0; i < n; i++ {
			go func(i int) {
				t.Log(fmt.Sprintf("put some_event/%d", i+1))
				j.PutEvent(fmt.Sprintf("some_event/%d", i+1), []byte{byte(i)})
			}(i)
		}
		wg := &sync.WaitGroup{}
		wg.Add(uNum)
		t.Log("run listeners ", time.Now())
		for user := 0; user < uNum; user++ {
			go func(user int) {
				tmpC := c
				var err error
				for i := 0; i < n; i++ {
					tmpC, err = j.WaitEvent(tmpC, "some_event/", context.Background())
					r.NoError(err)
					e, err := j.GetEventByCounter(tmpC)
					r.NoError(err)
					t.Log("user: ", user+1, " event: ", e)
				}
				wg.Done()
			}(user)
		}
		wg.Wait()
		t.Log("listeners finished ", time.Now())
		r.Equal(n, int(j.EventsCounter))
		r.Len(j.EventsUpdated, 0)
	})
}
