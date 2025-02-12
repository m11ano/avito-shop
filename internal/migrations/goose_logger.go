package migrations

import (
	"fmt"
	"log/slog"
)

type SlogLogger struct {
	logger *slog.Logger
}

func NewSlogLogger(logger *slog.Logger) *SlogLogger {
	return &SlogLogger{
		logger: logger,
	}
}

func (l *SlogLogger) Fatalf(format string, v ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, v...))
}

func (l *SlogLogger) Printf(format string, v ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, v...))
}
