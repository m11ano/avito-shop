package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/internal/app"
)

type InfoHandlerOut struct {
	Coins int64 `json:"coins"`
}

func (ctrl *Controller) InfoHandler(c *fiber.Ctx) error {
	if isAuth, ok := c.Locals("isAuth").(bool); !ok || !isAuth {
		return app.ErrUnauthorized
	}

	var accountID *uuid.UUID
	var ok bool
	if accountID, ok = c.Locals("authAccountID").(*uuid.UUID); !ok {
		return app.ErrUnauthorized
	}

	var err error
	out := InfoHandlerOut{}

	out.Coins, _, err = ctrl.usecaseOperation.GetBalanceByAccountID(c.Context(), *accountID)
	if err != nil {
		return err
	}

	return c.JSON(out)
}
