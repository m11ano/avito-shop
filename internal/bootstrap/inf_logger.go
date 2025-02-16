package bootstrap

import (
	"io"
	"log/slog"
	"os"

	"github.com/m11ano/avito-shop/internal/infra/config"
)

func NewLogger(config config.Config) *slog.Logger {
	var handler slog.Handler
	switch {
	case !config.App.UseLogger:
		handler = slog.NewTextHandler(io.Discard, nil)
	case config.App.IsProd:
		handler = slog.NewJSONHandler(os.Stdout, nil)
	default:
		handler = slog.NewTextHandler(os.Stdout, nil)
	}

	logger := slog.New(handler)
	return logger
}
