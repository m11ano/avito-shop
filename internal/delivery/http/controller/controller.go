package controller

import (
	"github.com/go-playground/validator/v10"
	"github.com/m11ano/avito-shop/internal/usecase"
)

type Controller struct {
	vldtr          *validator.Validate
	accountUsecase usecase.Account
}

func New(vldtr *validator.Validate, accountUsecase usecase.Account) *Controller {
	return &Controller{
		vldtr:          vldtr,
		accountUsecase: accountUsecase,
	}
}
