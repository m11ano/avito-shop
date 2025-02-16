package bootstrap

import (
	"log/slog"
	"os"

	"github.com/m11ano/avito-shop/internal/infra/config"
)

func NewLogger(config config.Config) *slog.Logger {
	var handler slog.Handler
	if !config.App.IsProd {
		handler = slog.NewTextHandler(os.Stdout, nil)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	}
	logger := slog.New(handler)
	return logger
}
