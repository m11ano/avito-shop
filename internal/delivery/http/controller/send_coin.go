package controller

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/internal/delivery/http/validation"
	"github.com/m11ano/avito-shop/internal/usecase"
	"github.com/m11ano/avito-shop/pkg/e"
)

type SendCoinIn struct {
	ToUser string `json:"toUser" validate:"required"`
	Amount int64  `json:"amount" validate:"gt=0"`
}

func (ctrl *Controller) SendCoinValidate(in *SendCoinIn) (bool, string) {
	if err := ctrl.vldtr.Struct(in); err != nil {
		return validation.FormatErrors(err)
	}
	return true, ""
}

func (ctrl *Controller) SendCoinHandler(c *fiber.Ctx) error {
	if isAuth, ok := c.Locals("isAuth").(bool); !ok || !isAuth {
		return e.ErrUnauthorized
	}

	in := &SendCoinIn{}

	if err := c.BodyParser(in); err != nil {
		return &fiber.Error{Code: fiber.ErrBadRequest.Code, Message: err.Error()}
	}

	var ok bool

	ok, errMsg := ctrl.SendCoinValidate(in)
	if !ok {
		return &fiber.Error{Code: fiber.ErrBadRequest.Code, Message: errMsg}
	}

	var accountID *uuid.UUID
	if accountID, ok = c.Locals("authAccountID").(*uuid.UUID); !ok {
		return e.ErrUnauthorized
	}

	var requestID *uuid.UUID
	if requestID, ok = c.Locals("requestID").(*uuid.UUID); !ok {
		return e.ErrInternal
	}

	_, _, err := ctrl.usecaseCoinTransfer.MakeTransferByUsername(c.Context(), in.ToUser, *accountID, in.Amount, requestID)
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
