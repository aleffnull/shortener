package logger

import "go.uber.org/zap"

type ZapLogger struct {
	sugar *zap.SugaredLogger
}

var _ Logger = (*ZapLogger)(nil)

func NewZapLogger(zap *zap.Logger) Logger {
	return &ZapLogger{
		sugar: zap.Sugar(),
	}
}

func (z *ZapLogger) Infof(template string, args ...any) {
	z.sugar.Infof(template, args...)
}

func (z *ZapLogger) Fatalf(template string, args ...any) {
	z.sugar.Fatalf(template, args...)
}
