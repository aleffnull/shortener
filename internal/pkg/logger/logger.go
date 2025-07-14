package logger

type Logger interface {
	Infof(template string, args ...any)
	Errorf(template string, args ...any)
	Fatalf(template string, args ...any)
}
