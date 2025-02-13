package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
)

func Recovery(logger *slog.Logger) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		defer func() {
			if errRec := recover(); errRec != nil {
				var err error
				switch errData := errRec.(type) {
				case error:
					err = errData
				case string:
					err = fmt.Errorf("panic: %s", errData)
				default:
					err = errors.New("panic: unknown error happend")
				}

				handlerErr := c.App().ErrorHandler(c, err)
				if handlerErr != nil {
					logger.ErrorContext(c.Context(), "failed to call fiber error handler", slog.Any("error", handlerErr), slog.Any("trackeback", string(debug.Stack())))
				}

				logger.ErrorContext(c.Context(), "panic inside http request", slog.Any("error", err), slog.Any("trackeback", string(debug.Stack())))
			}
		}()

		return c.Next()
	}
}
