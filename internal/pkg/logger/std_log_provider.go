package logger

import (
	"log"

	"go.uber.org/zap"
)

type StdLogProvider interface {
	GetStdLog() *log.Logger
}

type stdLogProviderImpl struct {
	stdLogger *log.Logger
}

var _ StdLogProvider = (*stdLogProviderImpl)(nil)

func NewStdLogProvider(zapLogger *zap.Logger) StdLogProvider {
	return &stdLogProviderImpl{
		stdLogger: zap.NewStdLog(zapLogger),
	}
}

func (i *stdLogProviderImpl) GetStdLog() *log.Logger {
	return i.stdLogger
}
