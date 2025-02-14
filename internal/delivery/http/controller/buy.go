package controller

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/internal/app"
	"github.com/m11ano/avito-shop/internal/usecase"
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
		// Маппинг ошибок под требования контракта
		switch {
		case errors.Is(err, usecase.ErrOperationNotEnoughFunds):
			return app.NewErrorFrom(app.ErrBadRequest).SetMessage(usecase.ErrOperationNotEnoughFunds.Message())
		case errors.Is(err, app.ErrConflict):
			return app.NewErrorFrom(app.ErrBadRequest).SetMessage(app.ErrConflict.Message())
		case errors.Is(err, app.ErrNotFound):
			return app.NewErrorFrom(app.ErrBadRequest).SetMessage(app.ErrNotFound.Message())
		default:
			return err
		}
	}

	return c.Status(fiber.StatusOK).SendString("")
}
