package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/m11ano/avito-shop/internal/app"
)

type errorJSON struct {
	Errors string `json:"errors"`
}

func ErrorHandler() func(*fiber.Ctx, error) error {
	return func(c *fiber.Ctx, err error) error {
		json := errorJSON{}
		code := 500

		switch errTyped := err.(type) {
		case *app.LogicError:
			code = errTyped.Code()
			json.Errors = errTyped.Error()
		case *fiber.Error:
			code = errTyped.Code
			json.Errors = err.Error()
		default:
			json.Errors = "internal error"
		}

		return c.Status(code).JSON(json)
	}
}
