package simpleLog

import (
	"fmt"
	"log"
	"os"
)

type SimpleLog struct {
	log *log.Logger
}

func NewSimpleLog() *SimpleLog {
	return &SimpleLog{log: log.New(os.Stdout, "", log.Ldate|log.Ltime)}
}

func (l *SimpleLog) Debug(v ...any) {
	f := "debug: " + fmt.Sprint(v...)
	l.log.Print(f)
}
func (l *SimpleLog) Debugf(format string, v ...any) {
	f := "debug: " + fmt.Sprintf(format, v...)
	l.log.Print(f)
}
func (l *SimpleLog) Info(v ...any) {
	f := "\ninfo: " + fmt.Sprintln(v...)
	l.log.Println(f)
}
func (l *SimpleLog) Infof(format string, v ...any) {
	f := "\ninfo: " + fmt.Sprintf(format, v...) + "\n"
	l.log.Println(f)
}
func (l *SimpleLog) Warn(v ...any) {
	f := "\nwarn: " + fmt.Sprintln(v...)
	l.log.Println(f)
}
func (l *SimpleLog) Warnf(format string, v ...any) {
	f := "\nwarn: " + fmt.Sprintf(format, v...) + "\n"
	l.log.Println(f)
}
func (l *SimpleLog) Error(v ...any) {
	f := "\nerror: " + fmt.Sprintln(v...)
	l.log.Println(f)
}
func (l *SimpleLog) Errorf(format string, v ...any) {
	f := "\nerror: " + fmt.Sprintf(format, v...) + "\n"
	l.log.Println(f)
}
func (l *SimpleLog) Fatal(v ...any) {
	l.log.Fatal(v...)
}
func (l *SimpleLog) Fatalf(format string, v ...any) {
	l.log.Fatalf(format, v...)
}
