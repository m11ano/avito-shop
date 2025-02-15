package middleware

import (
	"context"
	"log/slog"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type LogRequestData struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	IP        string `json:"ip"`
	URI       string `json:"uri"`
	RequestID string `json:"requestId,omitempty"`
}

func Logger(logger *slog.Logger) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		result := c.Next()

		if result != nil {
			err := c.App().ErrorHandler(c, result)
			if err != nil {
				logger.ErrorContext(c.Context(), "failed to call fiber error handler", slog.Any("error", err), slog.Any("trackeback", string(debug.Stack())))
			}
		}

		code := c.Response().StatusCode()

		logRequestData := LogRequestData{
			Method: c.Method(),
			Path:   c.Path(),
			IP:     c.IP(),
			URI:    string(c.Context().URI().QueryString()),
		}

		requestID := c.Locals("requestID")
		if requestID, ok := requestID.(*uuid.UUID); ok {
			logRequestData.RequestID = requestID.String()
		}

		go doLogging(c.Context(), logger, code, logRequestData)

		return nil
	}
}

func doLogging(ctx context.Context, logger *slog.Logger, code int, data LogRequestData) {
	if code >= 400 {
		logger.ErrorContext(
			ctx,
			"http response: error",
			slog.Int("reponseCode", code),
			slog.Any("request", data),
		)
	} else {
		logger.InfoContext(
			ctx,
			"http response: success",
			slog.Int("reponseCode", code),
			slog.Any("request", data),
		)
	}
}
