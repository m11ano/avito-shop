package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func TraceID() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		requestID, err := uuid.Parse(c.Get("X-Request-ID"))
		if err != nil {
			requestID = uuid.New()
		}

		c.Locals("requestID", &requestID)

		return c.Next()
	}
}
