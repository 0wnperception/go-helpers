package event_journal

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type EventJournalMock struct {
}

func NewEventJournalMock(basketSize int, persistent EventJournalPersistent) (j *EventJournalMock) {
	return &EventJournalMock{}
}

//put event to storage
func (j *EventJournalMock) PutEvent(label string, payload []byte) {
	fmt.Println("event put '", label, "' : ", string(payload))
}

//GetEventByNumber returns event with provided eventNumber if it's exist or error.
func (j *EventJournalMock) GetEventByNumber(eventNumber uint64) (e *Event, err error) {
	return &Event{
		ID:          uuid.New().String(),
		EventNumber: eventNumber,
		Time:        time.Now(),
		Label:       fmt.Sprintf("some event %d", e.EventNumber),
	}, nil
}

//GetNext returns next event if it exists or error.
func (j *EventJournalMock) GetNext(e *Event) (*Event, error) {
	return &Event{
		ID:          uuid.New().String(),
		EventNumber: e.EventNumber + 1,
		Time:        time.Now(),
		Label:       fmt.Sprintf("some event %d", e.EventNumber+1),
	}, nil
}

//FindByLabel tests whether existing events respond provided label.
func (j *EventJournalMock) FindByLabel(startEventNumber uint64, label string) (*Event, error) {
	return &Event{
		ID:          uuid.New().String(),
		EventNumber: startEventNumber + 5,
		Time:        time.Now(),
		Label:       label,
	}, nil
}
