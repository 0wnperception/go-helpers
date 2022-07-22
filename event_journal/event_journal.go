package event_journal

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type EventJournal struct {
	sync.Locker
	ID             string
	BasketsIDsList []string
	InMemoryBasket EventBasket
	Persistent     EventJournalPersistent
	BasketsCounter uint64
	EventsCounter  uint64
	EventsUpdated  map[string]chan struct{}
}

type EventBasket struct {
	ID                string
	Number            uint64
	StartEventCounter uint64
	Events            []Event
	Size              int
}

type EventCounter struct {
	Counter uint64
}

type Event struct {
	ID          string
	Time        time.Time
	EventNumber uint64
	Label       string
	Payload     []byte
}

type EventJournalPersistent interface {
	Store(*EventJournal) error
	Restore() (*EventJournal, error)
	GetEventByCounter(counter uint64) (*Event, error)
}

func NewEventJournal(basketSize int, persistent EventJournalPersistent) (j *EventJournal) {
	j = &EventJournal{
		Locker:        &sync.RWMutex{},
		ID:            uuid.New().String(),
		EventsUpdated: make(map[string]chan struct{}),
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

//puts event to storage
func (j *EventJournal) PutEvent(label string, payload []byte) {
	j.Lock()
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
		Payload:     payload,
	}
	//save event in memory and in persistent if it exists
	j.InMemoryBasket.Events[j.EventsCounter-j.InMemoryBasket.StartEventCounter-1] = e
	if j.Persistent != nil {
		if err := j.Persistent.Store(j); err != nil {
			panic(err)
		}
	}

	//close updaters
	for id, updated := range j.EventsUpdated {
		close(updated)
		delete(j.EventsUpdated, id)
	}
	j.Unlock()
}

//return new last event counter
func (j *EventJournal) GetEventCounter() EventCounter {
	return EventCounter{
		Counter: j.EventsCounter,
	}
}

//return event with event number provided with counter if it's exist or error
func (j *EventJournal) GetEventByCounter(counter EventCounter) (e *Event, err error) {
	if j.EventsCounter == 0 || counter.Counter > j.EventsCounter {
		return nil, errors.New("no events provided")
	}
	if counter.Counter > j.InMemoryBasket.StartEventCounter && counter.Counter <= j.InMemoryBasket.StartEventCounter+uint64(j.InMemoryBasket.Size) {
		tmp := j.InMemoryBasket.Events[counter.Counter-j.InMemoryBasket.StartEventCounter-1]
		return &tmp, nil
	} else {
		if j.Persistent != nil {
			return j.Persistent.GetEventByCounter(counter.Counter)
		} else {
			return nil, errors.New("no counter in memory, invalid persistent")
		}
	}
}

//return next counter if it exists or error and provided counter
func (j *EventJournal) GetNextCounter(counter EventCounter) (EventCounter, error) {
	if counter.Counter < j.EventsCounter {
		counter.Counter++
		return counter, nil
	} else {
		return counter, errors.New("no events provided")
	}
}

//return existing event after counter with label, if no new events provided waits for it
func (j *EventJournal) WaitEvent(counter EventCounter, label string, ctx context.Context) (EventCounter, error) {
	id := uuid.New().String()
	for {
		//check previous events
		if newCounter, ok := j.searchingForLabel(counter, label); ok {
			return newCounter, nil
		} else {
			counter = newCounter
		}
		//put waiting channel on event put
		ch := make(chan struct{})
		j.Lock()
		j.EventsUpdated[id] = ch
		j.Unlock()
		select {
		case <-ch:
		case <-ctx.Done():
			j.Lock()
			delete(j.EventsUpdated, id)
			j.Unlock()
			return counter, ctx.Err()
		}
	}
}

//goes through existing events started from counter to find label, if it is found return newCounter and ok=true
func (j *EventJournal) searchingForLabel(counter EventCounter, label string) (newCounter EventCounter, ok bool) {
	newCounter = counter
	var err error
	for {
		if newCounter, err = j.GetNextCounter(newCounter); err == nil {
			if j.checkEventLabel(newCounter, label) {
				ok = true
				return
			}
		} else {
			return
		}
	}
}

//checks event label for prefix by existing event number, it panics if it is not exist
func (j *EventJournal) checkEventLabel(counter EventCounter, label string) bool {
	if e, err := j.GetEventByCounter(counter); err != nil {
		panic("wrong job with event updated chan:" + err.Error())
	} else {
		return strings.HasPrefix(e.Label, label)
	}
}
