package event_log_service

import (
	"os"
)

type EventLoggerFile struct {
	f *os.File
}

func NewEventLoggerFile(path string) (*EventLoggerFile, error) {
	if f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err != nil {
		return nil, err
	} else {
		return &EventLoggerFile{f: f}, nil
	}
}

func (l *EventLoggerFile) Write(msg string) error {
	if _, err := l.f.Write([]byte(msg + "\n")); err != nil {
		return err
	}
	return nil
}
func (l *EventLoggerFile) Flush() error {
	return l.f.Close()
}
