package taskFlow

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPrimaryTask(t *testing.T) {
	r := require.New(t)

	t.Run("processing", func(t *testing.T) {
		v := 0
		pt := NewPrimaryTask("pt1", func(ctx context.Context) (err error) {
			v = 1
			return nil
		})
		tasks := pt.GetSubTasks()
		r.Len(tasks, 1)
		task := tasks[0]
		r.Equal(pt, task)

		ctx := context.Background()
		ready := make(chan error, 1)
		go task.Run(ctx, ready)
		err := <-ready

		r.Nil(err)
		t.Log(err)
		r.Equal(1, v)
	})
}

func TestBasicTask(t *testing.T) {
	r := require.New(t)

	t.Run("add primary tasks", func(t *testing.T) {
		v := 0
		bt := NewBasicTask("bt1", 2)
		bt.Add(NewPrimaryTask("pt1", func(ctx context.Context) (err error) {
			v += 1
			bt.SetDone("primary 1 done")
			return nil
		})).Add(NewPrimaryTask("pt2", func(ctx context.Context) (err error) {
			time.Sleep(time.Second)
			v += 1
			return nil
		}))
		tasks := bt.GetSubTasks()
		r.Len(tasks, 2)

		ctx := context.Background()
		ready := make(chan error, 1)
		for i := 0; i < 2; i++ {
			go tasks[i].Run(ctx, ready)
			r.Nil(<-ready)
			r.Equal(i+1, v)
			t.Log(<-bt.Done())
		}
	})
}

func TestBackground(t *testing.T) {
	r := require.New(t)

	t.Run("run", func(t *testing.T) {
		v := 0
		bt := NewBasicTask("bt", 2)
		bt1 := NewBasicTask("bt1", 1)
		bt1.Add(NewPrimaryTask("pt1", func(ctx context.Context) (err error) {
			v += 1
			bt1.SetDone("primary 1 done")
			return nil
		}))
		bt.Add(bt1)
		bt.Add(NewPrimaryTask("pt2", func(ctx context.Context) (err error) {
			v += 1
			d1 := <-bt1.Done()
			bt.SetDone("primary 2 done with " + d1.(string))
			return nil
		}))
		r.Len(bt.GetSubTasks(), 2)

		ctx := context.Background()
		ready := BackgroundFlow(ctx, NewConcurrentFlow(), bt)
		r.Nil(<-ready)
		t.Log(<-bt.Done())
	})

	t.Run("run with cancel", func(t *testing.T) {
		v := 0
		bt := NewBasicTask("bt", 2)
		bt1 := NewBasicTask("bt1", 1)
		bt1.Add(NewPrimaryTask("bt1 pt1", func(ctx context.Context) (err error) {
			v += 1
			bt1.SetDone("primary 1 done")
			return nil
		}))
		bt.Add(bt1)
		bt.Add(NewPrimaryTask("bt pt1", func(ctx context.Context) (err error) {
			v += 1
			d1 := <-bt1.Done()
			bt.SetDone("primary 2 done with " + d1.(string))
			return nil
		}))
		r.Len(bt.GetSubTasks(), 2)

		ctx, finish := context.WithCancel(context.Background())
		finish()
		ready := BackgroundFlow(ctx, NewConcurrentFlow(), bt)
		if ready != nil {
			r.Nil(<-ready)
		}
		r.Equal(0, v)
	})
}

func TestBackgroundParallel(t *testing.T) {
	r := require.New(t)

	t.Run("run parallel", func(t *testing.T) {
		v := 0

		bTask1 := NewBasicTask("bTask1", 2)
		bTask1.Add(NewPrimaryTask("pTask1", func(ctx context.Context) (err error) {
			time.Sleep(time.Second)
			v += 1
			bTask1.SetDone("task 1 sub 1 done")
			return nil
		})).Add(NewPrimaryTask("pTask2", func(ctx context.Context) (err error) {
			time.Sleep(time.Second)
			v += 1
			log.Println("bTask1 finished")
			return nil
		}))

		bTask2 := NewBasicTask("bTask2", 2)
		bTask2.Add(NewPrimaryTask("pTask1", func(ctx context.Context) (err error) {
			v += 1
			bTask2.SetDone("task 2 sub 1 done")
			return nil
		})).Add(NewPrimaryTask("pTask2", func(ctx context.Context) (err error) {
			time.Sleep(time.Millisecond)
			v += 1
			log.Println("bTask2 finished")
			return nil
		}))
		ctx := context.Background()
		ready := BackgroundParallelFlow(ctx, NewConcurrentFlow(), bTask1, bTask2)
		go func() {
			log.Println(<-bTask1.Done())
		}()
		go func() {
			log.Println(<-bTask2.Done())
		}()
		r.Nil(<-ready[0])
		r.Nil(<-ready[1])
		r.Equal(4, v)
	})
	t.Run("run parallel with cancel", func(t *testing.T) {
		v := 0

		bTask1 := NewBasicTask("bTask1", 2)
		bTask1.Add(NewPrimaryTask("pTask1", func(ctx context.Context) (err error) {
			time.Sleep(time.Second)
			v += 1
			bTask1.SetDone("task 1 sub 1 done")
			return nil
		})).Add(NewPrimaryTask("pTask2", func(ctx context.Context) (err error) {
			time.Sleep(time.Second)
			v += 1
			log.Println("bTask1 finished")
			return nil
		}))

		bTask2 := NewBasicTask("bTask2", 2)
		bTask2.Add(NewPrimaryTask("pTask1", func(ctx context.Context) (err error) {
			v += 1
			bTask2.SetDone("task 2 sub 1 done")
			return nil
		})).Add(NewPrimaryTask("pTask2", func(ctx context.Context) (err error) {
			time.Sleep(time.Millisecond)
			v += 1
			log.Println("bTask2 finished")
			return nil
		}))
		ctx, finish := context.WithCancel(context.Background())
		ready := BackgroundParallelFlow(ctx, NewConcurrentFlow(), bTask1, bTask2)
		go func() {
			log.Println(<-bTask1.Done())
		}()
		go func() {
			log.Println(<-bTask2.Done())
		}()
		time.Sleep(100 * time.Millisecond)
		finish()
		time.Sleep(time.Second)

		r.Nil(<-ready[0])
		r.Nil(<-ready[1])
		r.Equal(3, v)
	})
	t.Run("run parallel sync", func(t *testing.T) {
		v := 0

		syncBall := make(chan struct{})
		bTask1 := NewBasicTask("bTask1", 4)
		bTask1.Add(
			NewPrimaryTask("do job 1", func(ctx context.Context) (err error) {
				time.Sleep(time.Second)
				v += 1
				bTask1.SetDone("task 1 done")

				return nil
			})).Add(
			NewPrimaryTask("after job 1", func(ctx context.Context) (err error) {
				syncBall <- struct{}{}
				return nil
			})).Add(
			NewPrimaryTask("wait job 2", func(ctx context.Context) (err error) {
				<-syncBall
				return nil
			})).Add(
			NewPrimaryTask("do job 3", func(ctx context.Context) (err error) {
				time.Sleep(time.Millisecond)
				v += 1
				return nil
			}))

		bTask2 := NewBasicTask("bTask2", 4)
		bTask2.Add(
			NewPrimaryTask("wait job 1", func(ctx context.Context) (err error) {
				<-syncBall
				return nil
			})).Add(
			NewPrimaryTask("do job 2", func(ctx context.Context) (err error) {
				v += 1
				bTask2.SetDone("task 2 done")
				return nil
			})).Add(
			NewPrimaryTask("after job 2", func(ctx context.Context) (err error) {
				syncBall <- struct{}{}
				return nil
			})).Add(
			NewPrimaryTask("do job 4", func(ctx context.Context) (err error) {
				time.Sleep(time.Second)
				v += 1
				return nil
			}))
		ctx := context.Background()
		ready := BackgroundParallelFlow(ctx, nil, bTask1, bTask2)
		go func() {
			log.Println(<-bTask1.Done())
		}()
		go func() {
			log.Println(<-bTask2.Done())
		}()
		r.Nil(<-ready[0])
		r.Nil(<-ready[1])
		r.Equal(4, v)
	})
}
