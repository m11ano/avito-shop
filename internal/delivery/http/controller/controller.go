package controller

import (
	"github.com/go-playground/validator/v10"
	"github.com/m11ano/avito-shop/internal/usecase"
)

type Controller struct {
	vldtr               *validator.Validate
	usecaseAuth         usecase.Auth
	usecaseOperation    usecase.Operation
	usecaseShopPurchase usecase.ShopPurchase
	usecaseCoinTransfer usecase.CoinTransfer
}

func New(vldtr *validator.Validate, usecaseAuth usecase.Auth, usecaseOperation usecase.Operation, usecaseShopPurchase usecase.ShopPurchase, usecaseCoinTransfer usecase.CoinTransfer) *Controller {
	return &Controller{
		vldtr:               vldtr,
		usecaseAuth:         usecaseAuth,
		usecaseOperation:    usecaseOperation,
		usecaseShopPurchase: usecaseShopPurchase,
		usecaseCoinTransfer: usecaseCoinTransfer,
	}
}
