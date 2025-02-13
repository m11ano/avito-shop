package bootstrap

import (
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)
	return logger
}
