package simpleLog

import (
	"fmt"
	"log"
)

type SimpleLog struct {
}

func NewSimpleLog() *SimpleLog {
	return &SimpleLog{}
}

func (l *SimpleLog) Debug(v ...any) {
	f := "debug: " + fmt.Sprintln(v...)
	log.Println(f)
}
func (l *SimpleLog) Debugf(format string, v ...any) {
	f := "debug: " + format
	log.Printf(f, v...)
}
func (l *SimpleLog) Info(v ...any) {
	f := "info: " + fmt.Sprintln(v...)
	log.Println(f)
}
func (l *SimpleLog) Infof(format string, v ...any) {
	f := "info: " + format
	log.Printf(f, v...)
}
func (l *SimpleLog) Warn(v ...any) {
	f := "warn: " + fmt.Sprintln(v...)
	log.Println(f)
}
func (l *SimpleLog) Warnf(format string, v ...any) {
	f := "warn: " + format
	log.Printf(f, v...)
}
func (l *SimpleLog) Error(v ...any) {
	f := "error: " + fmt.Sprintln(v...)
	log.Println(f)
}
func (l *SimpleLog) Errorf(format string, v ...any) {
	f := "error: " + format
	log.Printf(f, v...)
}
func (l *SimpleLog) Fatal(v ...any) {
	log.Fatal(v...)
}
func (l *SimpleLog) Fatalf(format string, v ...any) {
	log.Fatalf(format, v...)
}
