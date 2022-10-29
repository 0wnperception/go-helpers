package event_log_service

import (
	"context"
	"fmt"

	"github.com/0wnperception/go-helpers/queue"

	"github.com/0wnperception/go-helpers/event_journal"
)

type EventLogServiceConfig struct {
	EventQueueSize int
}

type EventLoggerIface interface {
	Write(msg string) error
	Flush() error
}

type EventLogService struct {
	loggers      []EventLoggerIface
	journal      *event_journal.EventJournal
	eventsQ      queue.QueueIface[*event_journal.Event]
	writerActive bool
	writerDone   chan struct{}
	cancel       context.CancelFunc
	done         chan error
}

func NewEventLogService(cfg EventLogServiceConfig, journal *event_journal.EventJournal, loggers ...EventLoggerIface) *EventLogService {
	return &EventLogService{
		loggers: loggers,
		journal: journal,
		eventsQ: queue.NewQueue[*event_journal.Event](cfg.EventQueueSize),
		done:    make(chan error, 1),
	}
}

func (s *EventLogService) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go s.run(ctx)
	return nil
}

func (s *EventLogService) Stop() error {
	s.cancel()
	return nil
}

func (s *EventLogService) Done() <-chan error {
	return s.done
}

func (s *EventLogService) run(ctx context.Context) {
	var e *event_journal.Event
	waiter := s.journal.NewWaiter("/")
	for {
		surved := waiter.Surve(ctx)
		if ctx.Err() != nil {
			break
		} else {
			if surved != nil {
				s.runWriter(surved)
				e = surved
			} else {
				waiter = s.journal.NewWaiter("/")
			}
		}
	}
	if e != nil {
		var err error
		for {
			e, err = s.journal.GetNext(e)
			if err != nil {
				break
			} else {
				s.runWriter(e)
			}
		}
		<-s.writerDone
	}
	s.flushLoggers()
	close(s.done)
}

func (s *EventLogService) runWriter(e *event_journal.Event) {
	s.eventsQ.Push(e)
	if !s.writerActive {
		s.writerActive = true
		s.writerDone = make(chan struct{})
		go s.writer()
	}
}

func (s *EventLogService) writer() {
	for s.eventsQ.Len() > 0 {
		e, ok := s.eventsQ.Pull()
		if !ok {
			panic("wrong job with events queue")
		}
		s.writeEvent(e)
	}
	s.writerActive = false
	close(s.writerDone)
}

func (s *EventLogService) writeEvent(e *event_journal.Event) {
	var msg string
	if e.Payload != nil {
		msg = fmt.Sprintf("%d.(\"%s\").(\"%v\"): %v", e.EventNumber, e.Label, e.Time, string(e.Payload))
	} else {
		msg = fmt.Sprintf("%d.(\"%s\").(\"%v\")", e.EventNumber, e.Label, e.Time)
	}

	for _, l := range s.loggers {
		l.Write(msg)
	}
}

func (s *EventLogService) flushLoggers() {
	for _, l := range s.loggers {
		l.Flush()
	}
}
