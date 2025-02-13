package middleware

import "github.com/gofiber/fiber/v2"

type errorJSON struct {
	Errors string `json:"errors"`
}

func ErrorHandler() func(*fiber.Ctx, error) error {
	return func(c *fiber.Ctx, err error) error {
		json := errorJSON{}
		code := 500

		switch errTyped := err.(type) {
		// case app.Error:
		// 	json.Code = errTyped.LogicCode()
		// 	json.Details = errTyped.Details()
		// 	json.Error = errTyped.Error()
		case *fiber.Error:
			code = errTyped.Code
			json.Errors = err.Error()
		default:
			json.Errors = "internal error"
		}

		return c.Status(code).JSON(json)
	}
}
