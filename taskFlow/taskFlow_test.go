package taskFlow

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/0wnperception/go-helpers/logger/simple_log"

	"github.com/stretchr/testify/require"
)

func TestSimple(t *testing.T) {
	r := require.New(t)
	l := simple_log.NewSimpleLogger()
	t.Run("completed", func(t *testing.T) {
		v := 0
		id := "pt1"
		pt := New(id, 0).Task(func(ctx context.Context) (err error) {
			v = 1
			return nil
		})

		_, resChan := pt.Run(FlowRunConfig{Log: l,
			Ctx:                     context.Background(),
			FlowResultCodeCompleted: 0,
			FlowResultCodeWithError: 1,
		})
		res := <-resChan.NotifyAll()
		r.NotNil(res)
		r.Equal(1, v)
		r.Equal(Result{
			ID:         id,
			Code:       0,
			IsOriginal: true,
		}, *res)
	})
	t.Run("with error", func(t *testing.T) {
		v := 0
		id := "pt1"
		serr := errors.New("some err")
		task := func(ctx context.Context) (err error) {
			v = 1
			return serr
		}
		pt := New(id, 0).Task(task)

		_, resChan := pt.Run(FlowRunConfig{Log: l,
			Ctx:                     context.Background(),
			FlowResultCodeCompleted: 0,
			FlowResultCodeWithError: 1,
		})
		res := <-resChan.NotifyAll()
		r.NotNil(res)
		r.Equal(1, v)
		r.Equal(Result{
			ID:         id,
			Code:       1,
			IsOriginal: true,
			Payload:    serr,
		}, *res)
	})
}

func TestSeq(t *testing.T) {
	r := require.New(t)
	l := simple_log.NewSimpleLogger()
	t.Run("completed", func(t *testing.T) {
		v := 0
		task := func(ctx context.Context) (err error) {
			v++
			return nil
		}
		pt := New("pt1", 0).Task(task)
		pt.Next(New("pt2", 0)).Task(task).Next(New("pt3", 0).Task(task))

		_, resChan := pt.Run(FlowRunConfig{Log: l,
			Ctx:                     context.Background(),
			FlowResultCodeCompleted: 0,
			FlowResultCodeWithError: 1,
		})
		res := <-resChan.NotifyAll()
		r.NotNil(res)
		r.Equal(Result{
			ID:         "pt1",
			Code:       0,
			IsOriginal: true,
			Payload:    nil,
		}, *res)
		res = <-resChan.NotifyAll()
		r.NotNil(res)
		r.Equal(Result{
			ID:         "pt2",
			Code:       0,
			IsOriginal: true,
			Payload:    nil,
		}, *res)
		res = <-resChan.NotifyAll()
		r.NotNil(res)
		r.Equal(Result{
			ID:         "pt3",
			Code:       0,
			IsOriginal: true,
			Payload:    nil,
		}, *res)
		r.Equal(3, v)
	})
}

func TestSeqDone(t *testing.T) {
	r := require.New(t)
	l := simple_log.NewSimpleLogger()
	t.Run("completed", func(t *testing.T) {
		v := 0
		task := func(ctx context.Context) (err error) {
			v++
			return nil
		}
		pt := New("pt1", 0).Task(task)
		pt.Next(New("pt2", 0)).Task(task).Next(New("pt3", 0).Task(task))

		done, _ := pt.Run(FlowRunConfig{Log: l,
			Ctx:                     context.Background(),
			FlowResultCodeCompleted: 0,
			FlowResultCodeWithError: 1,
		})
		r.Nil(<-done)
		r.Equal(3, v)
	})
}

func TestParallel(t *testing.T) {
	r := require.New(t)
	l := simple_log.NewSimpleLogger()
	t.Run("completed", func(t *testing.T) {
		v, v1, v2, v3 := 0, 0, 0, 0
		flow := New("original", 3)
		sim1 := New("sim1", 0)
		task1 := func(ctx context.Context) (err error) {
			time.Sleep(time.Second)
			v1++
			sim1.FlowResult(FlowResultCode(3), v1)
			return nil
		}
		sim1.Task(task1).Next(New("sim1.2", 0).Task(task1)).Next(New("sim1.3", 0).Task(task1))

		sim2 := New("sim2", 0)
		task2 := func(ctx context.Context) (err error) {
			time.Sleep(time.Second + 20*time.Millisecond)
			v2++
			sim2.FlowResult(FlowResultCode(3), v2)
			return nil
		}
		sim2.Task(task2).Next(New("sim2.2", 0).Task(task2)).Next(New("sim2.3", 0).Task(task2))

		sim3 := New("sim3", 0)
		task3 := func(ctx context.Context) (err error) {
			time.Sleep(time.Second + 30*time.Millisecond)
			v3++
			sim3.FlowResult(FlowResultCode(3), v3)
			return nil
		}
		sim3.Task(task3).Next(New("sim3.2", 0)).Task(task3).Next(New("sim3.3", 0)).Task(task3)

		flow.Sub(sim1)
		flow.Sub(sim2)
		flow.Sub(sim3)
		flow.Next(
			New("flow 2", 0)).Task(
			func(ctx context.Context) (err error) {
				time.Sleep(time.Second)
				v++
				flow.FlowResult(FlowResultCode(3), v)
				return nil
			})

		_, resChan := flow.Run(FlowRunConfig{Log: l,
			Ctx:                     context.Background(),
			FlowResultCodeCompleted: 0,
			FlowResultCodeWithError: 1,
		})
		r.Equal(Result{
			ID:         "sim1",
			Code:       3,
			IsOriginal: false,
			Payload:    1,
		}, *<-resChan.NotifyAll())
		r.Equal(Result{
			ID:   "sim1",
			Code: 0,
		}, *<-resChan.NotifyAll())

		r.Equal(Result{
			ID:         "sim2",
			Code:       3,
			IsOriginal: false,
			Payload:    1,
		}, *<-resChan.NotifyAll())
		r.Equal(Result{
			ID:   "sim2",
			Code: 0,
		}, *<-resChan.NotifyAll())

		r.Equal(Result{
			ID:         "sim3",
			Code:       3,
			IsOriginal: false,
			Payload:    1,
		}, *<-resChan.NotifyAll())
		r.Equal(Result{
			ID:   "sim3",
			Code: 0,
		}, *<-resChan.NotifyAll())

		r.Equal(Result{
			ID:         "sim1",
			Code:       3,
			IsOriginal: false,
			Payload:    2,
		}, *<-resChan.NotifyAll())
		r.Equal(Result{
			ID:   "sim1.2",
			Code: 0,
		}, *<-resChan.NotifyAll())
		r.Equal(Result{
			ID:         "sim2",
			Code:       3,
			IsOriginal: false,
			Payload:    2,
		}, *<-resChan.NotifyAll())
		r.Equal(Result{
			ID:   "sim2.2",
			Code: 0,
		}, *<-resChan.NotifyAll())
		r.Equal(Result{
			ID:         "sim3",
			Code:       3,
			IsOriginal: false,
			Payload:    2,
		}, *<-resChan.NotifyAll())
		r.Equal(Result{
			ID:   "sim3.2",
			Code: 0,
		}, *<-resChan.NotifyAll())

		r.Equal(Result{
			ID:         "sim1",
			Code:       3,
			IsOriginal: false,
			Payload:    3,
		}, *<-resChan.NotifyAll())

		r.Equal(Result{
			ID: "sim1.3",
		}, *<-resChan.NotifyAll())

		r.Equal(Result{
			ID:      "sim2",
			Code:    3,
			Payload: 3,
		}, *<-resChan.NotifyAll())
		r.Equal(Result{
			ID: "sim2.3",
		}, *<-resChan.NotifyAll())

		r.Equal(Result{
			ID:         "sim3",
			Code:       3,
			IsOriginal: false,
			Payload:    3,
		}, *<-resChan.NotifyAll())
		r.Equal(Result{
			ID: "sim3.3",
		}, *<-resChan.NotifyAll())

		r.Equal(Result{
			ID:         "original",
			IsOriginal: true,
		}, *<-resChan.NotifyAll())
		r.Equal(Result{
			ID:         "original",
			Code:       3,
			IsOriginal: true,
			Payload:    1,
		}, *<-resChan.NotifyAll())

		r.Equal(Result{
			ID:         "flow 2",
			Code:       0,
			IsOriginal: true,
		}, *<-resChan.NotifyAll())

		r.Equal(1, v)
		r.Equal(3, v1)
		r.Equal(3, v2)
		r.Equal(3, v3)
	})
}
