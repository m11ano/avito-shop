package controller

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/m11ano/avito-shop/internal/usecase"
)

type Controller struct {
	logger              *slog.Logger
	vldtr               *validator.Validate
	usecaseAuth         usecase.Auth
	usecaseOperation    usecase.Operation
	usecaseShopPurchase usecase.ShopPurchase
	usecaseCoinTransfer usecase.CoinTransfer
}

func New(logger *slog.Logger, vldtr *validator.Validate, usecaseAuth usecase.Auth, usecaseOperation usecase.Operation, usecaseShopPurchase usecase.ShopPurchase, usecaseCoinTransfer usecase.CoinTransfer) *Controller {
	return &Controller{
		logger:              logger,
		vldtr:               vldtr,
		usecaseAuth:         usecaseAuth,
		usecaseOperation:    usecaseOperation,
		usecaseShopPurchase: usecaseShopPurchase,
		usecaseCoinTransfer: usecaseCoinTransfer,
	}
}
