package controller

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/internal/usecase"
	"github.com/m11ano/avito-shop/pkg/e"
)

func (ctrl *Controller) BuyHandler(c *fiber.Ctx) error {
	if isAuth, ok := c.Locals("isAuth").(bool); !ok || !isAuth {
		return e.ErrUnauthorized
	}

	var accountID *uuid.UUID
	var ok bool
	if accountID, ok = c.Locals("authAccountID").(*uuid.UUID); !ok {
		return e.ErrUnauthorized
	}

	var requestID *uuid.UUID
	if requestID, ok = c.Locals("requestID").(*uuid.UUID); !ok {
		return e.ErrInternal
	}

	shopItemName := c.Params("name")

	_, err := ctrl.usecaseShopPurchase.MakePurchase(c.Context(), shopItemName, *accountID, 1, requestID)
	if err != nil {
		// Маппинг ошибок под требования контракта
		switch {
		case errors.Is(err, usecase.ErrOperationNotEnoughFunds):
			return e.NewErrorFrom(e.ErrBadRequest).SetMessage(err.Error())
		case errors.Is(err, e.ErrConflict):
			return e.NewErrorFrom(e.ErrBadRequest).SetMessage(err.Error())
		case errors.Is(err, e.ErrNotFound):
			return e.NewErrorFrom(e.ErrBadRequest).SetMessage(err.Error())
		default:
			return err
		}
	}

	return c.SendStatus(fiber.StatusOK)
}
