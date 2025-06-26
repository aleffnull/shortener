package logger

import "context"

type Logger interface {
	Infof(template string, args ...any)
	Fatalf(template string, args ...any)
}

type ctxLoggerKey string

const loggerKey ctxLoggerKey = "logger"

func ContextWithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func LoggerFromContext(ctx context.Context) Logger {
	log, ok := ctx.Value(loggerKey).(Logger)
	if !ok {
		return NewDefaultLogger()
	}

	return log
}
