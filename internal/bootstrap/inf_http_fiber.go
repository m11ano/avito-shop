package bootstrap

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/m11ano/avito-shop/internal/delivery/http/middleware"
)

type HTTPConfig struct {
	UnderProxy bool
	UseTraceID bool
	UseLogger  bool
}

func NewHTTPFiber(httpCfg HTTPConfig, logger *slog.Logger) *fiber.App {
	fiberCfg := fiber.Config{
		ErrorHandler: middleware.ErrorHandler(),
	}

	if httpCfg.UnderProxy {
		fiberCfg.ProxyHeader = fiber.HeaderXForwardedFor
	}

	app := fiber.New(fiberCfg)

	app.Use(middleware.Recovery(logger))

	if httpCfg.UseTraceID {
		app.Use(middleware.TraceID())
	}

	if httpCfg.UseLogger {
		app.Use(middleware.Logger(logger))
	}

	return app
}
