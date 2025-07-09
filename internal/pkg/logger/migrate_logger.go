package logger

import "github.com/golang-migrate/migrate/v4"

type MigrateLogger struct {
	logger Logger
}

var _ migrate.Logger = (*MigrateLogger)(nil)

func NewMigrateLogger(logger Logger) migrate.Logger {
	return &MigrateLogger{
		logger: logger,
	}
}

func (l *MigrateLogger) Printf(format string, args ...any) {
	l.logger.Infof(format, args...)
}

func (*MigrateLogger) Verbose() bool {
	return true
}
