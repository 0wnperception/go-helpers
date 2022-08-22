package event_journal

import (
	"context"
	"fmt"
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
		e, err := j.GetEventByNumber(1)
		r.NoError(err)
		t.Log(e)
		t.Log(j.InMemoryBasket.Events)
	})

	t.Run("one wait", func(t *testing.T) {
		j := NewEventJournal(4, nil)
		j.PutEvent("some_event/1", []byte{123})
		waiter := j.NewWaiter("")
		go func() {
			for i := 0; i < 3; i++ {
				time.Sleep(time.Millisecond)
				t.Log(fmt.Sprintf("put some_event/%d", i+1))
				j.PutEvent(fmt.Sprintf("some_event/%d", i+1), []byte{byte(i)})
			}
		}()
		for i := 0; i < 3; i++ {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
			e := waiter.Surve(ctx)
			r.NotNil(e)
			r.Equal([]byte{byte(i)}, e.Payload)
			t.Log(e)
			cancel()
		}
	})

	t.Run("many waiters", func(tMW *testing.T) {
		n := 100
		uNum := 5
		j := NewEventJournal(n, nil)

		tMW.Run("parallel", func(tP *testing.T) {
			tP.Log("run listeners ", time.Now())
			for user := 0; user < uNum; user++ {
				tP.Run(fmt.Sprintf("run parallel listener %d", user), func(tL *testing.T) {
					num := user
					tL.Parallel()
					waiter := j.NewWaiter("")
					for i := 0; i < n; i++ {
						ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
						e := waiter.Surve(ctx)
						r.NotNil(e)
						r.Equal([]byte{byte(i)}, e.Payload)
						tL.Log("listener: ", num, " event: ", e)
						cancel()
					}
				})
			}
			tP.Log("run providers ", time.Now())
			tP.Run(fmt.Sprintf("run parallel provider"), func(tPR *testing.T) {
				tPR.Parallel()
				time.Sleep(time.Millisecond)
				for i := 0; i < n; i++ {
					tPR.Log(fmt.Sprintf("put some_event/%d", i+1))
					j.PutEvent(fmt.Sprintf("some_event/%d", i+1), []byte{byte(i)})
					time.Sleep(time.Millisecond)
				}
				r.Equal(n, int(j.EventsCounter))
			})
		})

		r.Len(j.Waiters, uNum)
		for i := 0; i <= WAITER_BUCKET_SIZE; i++ {
			j.PutEvent("zero", []byte{})
		}
		r.Len(j.Waiters, 0)
	})

	t.Run("label check", func(tMW *testing.T) {
		j := NewEventJournal(4, nil)
		label := "some_event"
		waiter := j.NewWaiter(label)

		j.PutEvent("koko/2", []byte{211})
		j.PutEvent(label+"/1", []byte{123})

		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
		e := waiter.Surve(ctx)
		r.Equal(label+"/1", e.Label)
		cancel()
	})

	t.Run("label check context", func(tMW *testing.T) {
		j := NewEventJournal(1, nil)
		waiter := j.NewWaiter("")
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Millisecond))
		e := waiter.Surve(ctx)
		r.Nil(e)
		r.Equal(context.DeadlineExceeded, ctx.Err())
		cancel()
	})
}

func TestSeveralLabels(t *testing.T) {
	r := require.New(t)
	j := NewEventJournal(4, nil)
	waiter := j.NewWaiter("some_event/1", "some_event/2")
	j.PutEvent("bad_event/1", nil)
	j.PutEvent("some_event/1", []byte{1})
	j.PutEvent("some_event/2", []byte{2})
	j.PutEvent("some_event/3", []byte{3})
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
	for i := 0; i < 2; i++ {
		e := waiter.Surve(ctx)
		if e == nil {
			t.Log("context finished")
		}
		r.NotNil(e)
		r.Equal(fmt.Sprintf("some_event/%d", i+1), e.Label)
		t.Log(e)
	}
	cancel()
}
