package event_log_service

import "fmt"

type EventLoggerStd struct {
}

func NewEventLoggerStd() *EventLoggerStd {
	return &EventLoggerStd{}
}

func (l *EventLoggerStd) Write(msg string) error {
	fmt.Println(msg)
	return nil
}
func (l *EventLoggerStd) Flush() error {
	return nil
}
