package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/m11ano/avito-shop/internal/app"
)

type InfoHandlerOut struct{}

func (ctrl *Controller) InfoHandler(c *fiber.Ctx) error {
	if isAuth, ok := c.Locals("isAuth").(bool); !ok || !isAuth {
		return app.ErrUnauthorized
	}

	out := InfoHandlerOut{}

	return c.JSON(out)
}
