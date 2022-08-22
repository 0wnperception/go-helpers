package event_journal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const WAITER_BUCKET_SIZE = 100

type EventJournalIface interface {
	PutEvent(label string, payload interface{}) error
	GetEventByNumber(eventNumber uint64) (e *Event, err error)
	GetNext(e *Event) (*Event, error)
	FindByLabel(startEventNumber uint64, label string) (*Event, error)
	NewWaiter(label string) (w *Waiter)
}

type EventJournal struct {
	ID             string
	putLocker      sync.Locker
	waitersLocker  sync.Locker
	BasketsIDsList []string
	InMemoryBasket EventBasket
	Persistent     EventJournalPersistent
	BasketsCounter uint64
	EventsCounter  uint64
	Waiters        map[string]*Waiter
}

type Waiter struct {
	id     string
	ch     chan *Event
	labels []string
}

func (w *Waiter) Surve(ctx context.Context) *Event {
	for {
		select {
		case e := <-w.ch:
			if e != nil {
				for _, label := range w.labels {
					if e.IsEvent(label) {
						return e
					}
				}
			} else {
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}

type EventBasket struct {
	ID                string
	Number            uint64
	StartEventCounter uint64
	Events            []Event
	Size              int
}

type Event struct {
	ID          string    `json:"id"`
	Time        time.Time `json:"time"`
	EventNumber uint64    `json:"event_number"`
	Label       string    `json:"label"`
	Payload     []byte    `json:"payload"`
}

func (e *Event) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID          string `json:"id"`
		Time        string `json:"time"`
		EventNumber uint64 `json:"event_number"`
		Label       string `json:"label"`
		Payload     string `json:"payload"`
	}{
		ID:          e.ID,
		Time:        e.Time.String(),
		EventNumber: e.EventNumber,
		Label:       e.Label,
		Payload:     string(e.Payload),
	})
}

func (e *Event) GetEventNumber() uint64 {
	return e.EventNumber
}

func (e *Event) UnmarshalPayload(target interface{}) error {
	return json.Unmarshal(e.Payload, target)
}

//IsEvent tests whether the event label responds provided label.
func (e *Event) IsEvent(label string) bool {
	return strings.HasPrefix(e.Label, label)
}

type EventJournalPersistent interface {
	Store(*EventJournal) error
	Restore() (*EventJournal, error)
	GetEventByNumber(counter uint64) (*Event, error)
}

func NewEventJournal(basketSize int, persistent EventJournalPersistent) (j *EventJournal) {
	j = &EventJournal{
		ID:            uuid.New().String(),
		putLocker:     &sync.RWMutex{},
		waitersLocker: &sync.RWMutex{},
		Waiters:       make(map[string]*Waiter),
	}
	j.initBasket(basketSize)
	return
}

func EventJournalFromPersistent(persistent EventJournalPersistent) (j *EventJournal, err error) {
	if persistent != nil {
		return persistent.Restore()
	} else {
		return nil, errors.New("invalid persistent")
	}
}

func (j *EventJournal) initBasket(size int) {
	id := uuid.New().String()
	j.BasketsIDsList = append(j.BasketsIDsList, id)
	j.BasketsCounter++
	j.InMemoryBasket = EventBasket{
		ID:                id,
		Number:            j.BasketsCounter,
		StartEventCounter: j.EventsCounter,
		Events:            make([]Event, size),
		Size:              size,
	}
}

//flushes basket to persistent and inits new basket with the same size
func (j *EventJournal) flushBasket() error {
	if j.Persistent != nil {
		if err := j.Persistent.Store(j); err != nil {
			return err
		}
	}
	j.initBasket(j.InMemoryBasket.Size)
	return nil
}

//put event to storage
func (j *EventJournal) PutEvent(label string, payload interface{}) (err error) {
	var bin []byte
	if payload != nil {
		bin, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}
	j.putLocker.Lock()
	//check in memory size if it's full, flushes basket to persistent and inits new basket with the same size
	if len(j.InMemoryBasket.Events) == int(j.EventsCounter-j.InMemoryBasket.StartEventCounter) {
		if err := j.flushBasket(); err != nil {
			panic(err)
		}
	}
	id := uuid.New().String()
	//increase event counter and creates new one
	j.EventsCounter++
	e := Event{
		ID:          id,
		Time:        time.Now(),
		EventNumber: j.EventsCounter,
		Label:       label,
		Payload:     bin,
	}
	//save event in memory and in persistent if it exists
	j.InMemoryBasket.Events[j.EventsCounter-j.InMemoryBasket.StartEventCounter-1] = e
	if j.Persistent != nil {
		if err := j.Persistent.Store(j); err != nil {
			panic(err)
		}
	}

	//notify waiters
	j.waitersLocker.Lock()
	for _, waiter := range j.Waiters {
		j.notifyWaiter(e, waiter)
	}
	j.waitersLocker.Unlock()
	j.putLocker.Unlock()
	return nil
}

func (j *EventJournal) NewWaiter(labels ...string) (w *Waiter) {
	w = &Waiter{
		id:     uuid.New().String(),
		ch:     make(chan *Event, WAITER_BUCKET_SIZE),
		labels: labels,
	}
	j.waitersLocker.Lock()
	j.Waiters[w.id] = w
	j.waitersLocker.Unlock()
	return
}

func (j *EventJournal) notifyWaiter(e Event, waiter *Waiter) {
	if cap(waiter.ch) == len(waiter.ch) {
		close(waiter.ch)
		delete(j.Waiters, waiter.id)
	} else {
		waiter.ch <- &e
	}
}

//GetEventByNumber returns event with provided eventNumber if it's exist or error.
func (j *EventJournal) GetEventByNumber(eventNumber uint64) (e *Event, err error) {
	if j.EventsCounter == 0 || eventNumber > j.EventsCounter {
		return nil, errors.New(fmt.Sprintf("invalid number %d", eventNumber))
	}
	if eventNumber > j.InMemoryBasket.StartEventCounter && eventNumber <= j.InMemoryBasket.StartEventCounter+uint64(j.InMemoryBasket.Size) {
		tmp := j.InMemoryBasket.Events[eventNumber-j.InMemoryBasket.StartEventCounter-1]
		return &tmp, nil
	} else {
		if j.Persistent != nil {
			return j.Persistent.GetEventByNumber(eventNumber)
		} else {
			return nil, errors.New("no counter in memory, invalid persistent")
		}
	}
}

//GetNext returns next event if it exists or error.
func (j *EventJournal) GetNext(e *Event) (*Event, error) {
	return j.GetEventByNumber(e.EventNumber + 1)
}

//FindByLabel tests whether existing events respond provided label.
func (j *EventJournal) FindByLabel(startEventNumber uint64, label string) (*Event, error) {
	var (
		tmp *Event
		err error
	)
	for i := startEventNumber; i < j.EventsCounter; i++ {
		if i > j.InMemoryBasket.StartEventCounter && i <= j.InMemoryBasket.StartEventCounter+uint64(j.InMemoryBasket.Size) {
			e := j.InMemoryBasket.Events[i-j.InMemoryBasket.StartEventCounter-1]
			tmp = &e
		} else {
			if j.Persistent != nil {
				tmp, err = j.Persistent.GetEventByNumber(i)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("no counter in memory, invalid persistent")
			}
		}
		if tmp.IsEvent(label) {
			return tmp, nil
		}
	}
	return nil, errors.New("no events responding such label")
}
