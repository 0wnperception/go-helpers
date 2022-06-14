package simple_log

import (
	"fmt"
	"log"
	"os"
)

type SimpleLogger struct {
	log *log.Logger
}

func NewSimpleLogger() *SimpleLogger {
	return &SimpleLogger{log: log.New(os.Stdout, "", log.Ldate|log.Ltime)}
}

func (l *SimpleLogger) Debug(v ...any) {
	f := "debug: " + fmt.Sprint(v...)
	l.log.Print(f)
}
func (l *SimpleLogger) Debugf(format string, v ...any) {
	f := "debug: " + fmt.Sprintf(format, v...)
	l.log.Print(f)
}
func (l *SimpleLogger) Info(v ...any) {
	f := "\ninfo: " + fmt.Sprintln(v...)
	l.log.Println(f)
}
func (l *SimpleLogger) Infof(format string, v ...any) {
	f := "\ninfo: " + fmt.Sprintf(format, v...) + "\n"
	l.log.Println(f)
}
func (l *SimpleLogger) Warn(v ...any) {
	f := "\nwarn: " + fmt.Sprintln(v...)
	l.log.Println(f)
}
func (l *SimpleLogger) Warnf(format string, v ...any) {
	f := "\nwarn: " + fmt.Sprintf(format, v...) + "\n"
	l.log.Println(f)
}
func (l *SimpleLogger) Error(v ...any) {
	f := "\nerror: " + fmt.Sprintln(v...)
	l.log.Println(f)
}
func (l *SimpleLogger) Errorf(format string, v ...any) {
	f := "\nerror: " + fmt.Sprintf(format, v...) + "\n"
	l.log.Println(f)
}
func (l *SimpleLogger) Fatal(v ...any) {
	l.log.Fatal(v...)
}
func (l *SimpleLogger) Fatalf(format string, v ...any) {
	l.log.Fatalf(format, v...)
}
