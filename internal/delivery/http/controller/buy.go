package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/internal/app"
)

func (ctrl *Controller) BuyHandler(c *fiber.Ctx) error {
	if isAuth, ok := c.Locals("isAuth").(bool); !ok || !isAuth {
		return app.ErrUnauthorized
	}

	var accountID *uuid.UUID
	var ok bool
	if accountID, ok = c.Locals("authAccountID").(*uuid.UUID); !ok {
		return app.ErrUnauthorized
	}

	var requestID *uuid.UUID
	if requestID, ok = c.Locals("requestID").(*uuid.UUID); !ok {
		return app.ErrInternal
	}

	shopItemName := c.Params("name")

	_, err := ctrl.usecaseShopPurchase.MakePurchase(c.Context(), shopItemName, *accountID, 1, requestID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).SendString("")
}
