package logger

import "log"

type DefaultLogger struct{}

var _ Logger = (*DefaultLogger)(nil)

func NewDefaultLogger() Logger {
	return &DefaultLogger{}
}

func (*DefaultLogger) Infof(template string, args ...any) {
	log.Printf(template, args...)
}

func (*DefaultLogger) Fatalf(template string, args ...any) {
	log.Fatalf(template, args...)
}
