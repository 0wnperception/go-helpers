package taskFlow

import (
	"context"
	"sync"
)

type Logger interface {
	Debug(v ...any)
	Debugf(format string, v ...any)
	Info(v ...any)
	Infof(format string, v ...any)
	Warn(v ...any)
	Warnf(format string, v ...any)
	Error(v ...any)
	Errorf(format string, v ...any)
	Fatal(v ...any)
	Fatalf(format string, v ...any)
}

type FlowResultCode int

const (
	ORIGINAL_RESULTS_CAP = 3
)

type Results struct {
	locker          sync.Locker
	notifierStarted bool
	notifyAll       bool
	notifierStoped  bool
	notifications   map[string]map[FlowResultCode]chan *Result
	results         chan *Result
}

type Result struct {
	ID         string
	Code       FlowResultCode
	IsOriginal bool
	Payload    interface{}
}

func newResults() *Results {
	return &Results{
		locker:        &sync.RWMutex{},
		notifications: make(map[string]map[FlowResultCode]chan *Result),
		results:       make(chan *Result, ORIGINAL_RESULTS_CAP),
	}
}

func (r *Results) NotifyAll() (chResult <-chan *Result) {
	if !r.notifierStarted {
		r.notifyAll = true
		chResult = r.results
	}
	return
}

func (r *Results) NotifyWith(code FlowResultCode, ID string) (chResult <-chan *Result) {
	if !r.notifyAll {
		if n, ok := r.notifications[ID]; ok {
			if ch, ok := n[code]; ok {
				chResult = ch
			} else {
				ch := make(chan *Result, 1)
				r.locker.Lock()
				n[code] = ch
				r.locker.Unlock()
				chResult = ch
			}
		} else {
			r.notifications[ID] = make(map[FlowResultCode]chan *Result)
			ch := make(chan *Result, 1)
			r.notifications[ID][code] = ch
			chResult = ch
		}

		if !r.notifierStarted {
			r.notifierStarted = true
			go r.startNotifier()
		}
	}
	return
}

func (r *Results) startNotifier() {
	for {
		select {
		case res := <-r.results:
			if res != nil {
				if n, ok := r.notifications[res.ID]; ok {
					if ch, ok := n[res.Code]; ok {
						if len(ch) == cap(ch) {
							<-ch
						}
						ch <- res
					}
				}
			} else {
				return
			}
		}
	}
}

func (r *Results) stopNotifier() {
	r.locker.Lock()
	r.notifierStoped = true
	close(r.results)
	r.locker.Unlock()
}

func (r *Results) setResult(result *Result) {
	if !r.notifierStoped {
		r.locker.Lock()
		if len(r.results) == cap(r.results) {
			<-r.results
		}
		r.results <- result
		r.locker.Unlock()
	}
}

type FlowConfig struct {
}

type Flow struct {
	id               string
	isOriginal       bool
	task             func(ctx context.Context) (err error)
	subFlowsAmount   int
	subFlowsCapacity int
	subFlows         []*Flow
	nextFlow         *Flow
	results          *Results
}

func New(id string, subFlowsCapacity int) *Flow {
	var subFlows []*Flow
	if subFlowsCapacity > 0 {
		subFlows = make([]*Flow, subFlowsCapacity)
	}
	return &Flow{
		id:               id,
		subFlowsCapacity: subFlowsCapacity,
		subFlows:         subFlows,
	}
}

func (f *Flow) Sub(subFlow *Flow) *Flow {
	if f.subFlowsCapacity > 0 {
		f.subFlows[f.subFlowsAmount] = subFlow
		f.subFlowsCapacity--
		f.subFlowsAmount++
		return subFlow
	} else {
		return nil
	}
}

func (f *Flow) Next(nextFlow *Flow) (next *Flow) {
	f.nextFlow = nextFlow
	return f.nextFlow
}

func (f *Flow) Task(task func(ctx context.Context) (err error)) *Flow {
	f.task = task
	return f
}

type FlowRunConfig struct {
	Ctx                     context.Context
	Log                     Logger
	FlowResultCodeCompleted FlowResultCode
	FlowResultCodeWithError FlowResultCode
}

func (f *Flow) Run(cfg FlowRunConfig) (done chan error, results *Results) {
	results = newResults()
	d := make(chan error, 1)
	done = d
	f.setResults(results)
	f.setOriginal(true)
	go f.run(cfg, d)
	return
}

func (f *Flow) setResults(results *Results) {
	f.results = results
}

func (f *Flow) setOriginal(state bool) {
	f.isOriginal = state
}

func (f *Flow) run(cfg FlowRunConfig, done chan error) {
	if cfg.Log != nil {
		cfg.Log.Infof("run flow '%s'", f.id)
	}
	if err := f.processTask(cfg); err != nil {
		if cfg.Log != nil {
			cfg.Log.Errorf("finish flow '%s' with error '%s'", f.id, err.Error())
		}
		f.FlowResult(cfg.FlowResultCodeWithError, err)
		done <- err
		if f.isOriginal {
			f.results.stopNotifier()
			close(done)
		}
	} else {
		f.processSubFlows(cfg)
		if cfg.Log != nil {
			cfg.Log.Infof("finish flow '%s'", f.id)
		}
		f.FlowResult(cfg.FlowResultCodeCompleted, nil)

		if cfg.Ctx.Err() == nil && f.nextFlow != nil {
			f.nextFlow.setResults(f.results)
			f.nextFlow.setOriginal(f.isOriginal)
			go f.nextFlow.run(cfg, done)
		} else {
			done <- cfg.Ctx.Err()
			if f.isOriginal {
				f.results.stopNotifier()
				close(done)
			}
		}
	}
}

func (f *Flow) processTask(cfg FlowRunConfig) error {
	if f.task != nil {
		if err := f.task(cfg.Ctx); err != nil {
			return err
		}
	}
	return nil
}

func (f *Flow) processSubFlows(cfg FlowRunConfig) {
	if len(f.subFlows) > 0 {
		done := make(chan error, len(f.subFlows))
		for subID := 0; subID < f.subFlowsAmount; subID++ {
			if cfg.Ctx.Err() == nil {
				f.subFlows[subID].setOriginal(false)
				f.subFlows[subID].setResults(f.results)
				go f.subFlows[subID].run(cfg, done)
			}
		}
		for subID := 0; subID < f.subFlowsAmount; subID++ {
			if cfg.Ctx.Err() == nil {
				select {
				case <-done:
				case <-cfg.Ctx.Done():
				}
			} else {
				break
			}
		}
	}
}

func (f *Flow) FlowResult(code FlowResultCode, payload interface{}) {
	if f.results != nil {
		f.results.setResult(&Result{
			ID:         f.id,
			IsOriginal: f.isOriginal,
			Code:       code,
			Payload:    payload,
		})
	}
}
