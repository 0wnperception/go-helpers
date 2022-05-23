package taskFlow

import (
	"context"
	"log"
)

type BasicTask struct {
	title    string
	done     chan interface{}
	amount   int
	capacity int
	subTasks []Task
}

type Task interface {
	GetTitle() string
	GetSubTasks() []*PrimaryTask
}

type PrimaryTask struct {
	title string
	task  func(ctx context.Context) (err error)
}

type ReadyPair struct {
	number int
	err    error
}

func NewPrimaryTask(title string, task func(ctx context.Context) (err error)) *PrimaryTask {
	return &PrimaryTask{
		title: title,
		task:  task,
	}
}

func (t *PrimaryTask) Run(ctx context.Context, ready chan<- error) {
	ready <- t.task(ctx)
}

func (t *PrimaryTask) RunParallel(ctx context.Context, ready chan<- *ReadyPair, number int) {
	ready <- &ReadyPair{
		err:    t.task(ctx),
		number: number,
	}
}

func (t *PrimaryTask) GetSubTasks() []*PrimaryTask {
	return []*PrimaryTask{t}
}

func (t *PrimaryTask) GetTitle() string {
	return t.title
}

func NewBasicTask(title string, capacity int) *BasicTask {
	return &BasicTask{
		title:    title,
		done:     make(chan interface{}, 1),
		capacity: capacity,
		subTasks: make([]Task, capacity),
	}
}

func (t *BasicTask) Add(task Task) *BasicTask {
	if t.amount < t.capacity {
		t.subTasks[t.amount] = task
		t.amount++
	}
	return t
}

func (t *BasicTask) GetSubTasks() (subTasks []*PrimaryTask) {
	for _, task := range t.subTasks {
		subTasks = append(subTasks, task.GetSubTasks()...)
	}
	return
}

func (t *BasicTask) SetDone(done interface{}) {
	if len(t.done) > 0 {
		<-t.done
	}
	t.done <- done
	close(t.done)
}
func (t *BasicTask) Done() <-chan interface{} {
	return t.done
}

func (t *BasicTask) GetTitle() string {
	return t.title
}

func BackgroundFlow(ctx context.Context, t *BasicTask) (done <-chan interface{}, ready <-chan error) {
	r := make(chan error, 1)
	go runBackgroundFlow(ctx, t, r)
	return t.Done(), r
}

func runBackgroundFlow(ctx context.Context, t *BasicTask, ready chan<- error) {
	locReady := make(chan error, 1)
	jobs := t.GetSubTasks()
JOBS:
	for _, job := range jobs {
		go job.Run(ctx, locReady)
		select {
		case err := <-locReady:
			if err != nil {
				ready <- err
				break JOBS
			}
		case <-ctx.Done():
			break JOBS
		}
	}
	close(ready)
	return
}

func BackgroundParallelFlow(ctx context.Context, bTasks ...*BasicTask) (done []<-chan interface{}, ready []<-chan error) {
	done = make([]<-chan interface{}, len(bTasks))
	ready = make([]<-chan error, len(bTasks))
	r := make([]chan error, len(bTasks))
	for i, bt := range bTasks {
		done[i] = bt.Done()
		tmpReady := make(chan error, 1)
		ready[i] = tmpReady
		r[i] = tmpReady
	}
	go runBackgroundParallelFlow(ctx, bTasks, r)
	return
}

func runBackgroundParallelFlow(ctx context.Context, bTasks []*BasicTask, ready []chan error) {
	if len(bTasks) > 0 {
		locReady := make(chan *ReadyPair, len(bTasks))
		counterArr := make([]int, len(bTasks))
		subTasksArr := make([][]*PrimaryTask, len(bTasks))

		for idx, job := range bTasks {
			subTasksArr[idx] = job.GetSubTasks()
			locReady <- &ReadyPair{
				err:    nil,
				number: idx,
			}
		}
	PROCESSING:
		for {
			select {
			case rp := <-locReady:
				if rp.err != nil {
					counterArr[rp.number] = -1
					ready[rp.number] <- rp.err
				} else {
					if counterArr[rp.number] < len(subTasksArr[rp.number]) {
						log.Printf("run %s %s", bTasks[rp.number].GetTitle(), subTasksArr[rp.number][counterArr[rp.number]].GetTitle())
						go subTasksArr[rp.number][counterArr[rp.number]].RunParallel(ctx, locReady, rp.number)
						counterArr[rp.number]++
					} else {
						counterArr[rp.number] = -1
						finish := true
						for _, counter := range counterArr {
							if counter != -1 {
								finish = false
								break
							}
						}
						if finish {
							break PROCESSING
						}
					}
				}
			case <-ctx.Done():
				break PROCESSING
			}
		}
	}
	for _, r := range ready {
		close(r)
	}
	return
}
