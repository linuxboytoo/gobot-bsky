package gobotbsky

import "fmt"

type logger interface {
	Debug(msg string)
}

type NoopLogger struct{}

func (l *NoopLogger) Debug(msg string) {}

type SimpleLogger struct{}

func (l *SimpleLogger) Debug(msg string) {
	fmt.Println(msg)
}
