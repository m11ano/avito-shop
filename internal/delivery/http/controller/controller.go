package controller

import (
	"github.com/go-playground/validator/v10"
	"github.com/m11ano/avito-shop/internal/usecase"
)

type Controller struct {
	vldtr            *validator.Validate
	usecaseAuth      usecase.Auth
	usecaseOperation usecase.Operation
}

func New(vldtr *validator.Validate, usecaseAuth usecase.Auth, usecaseOperation usecase.Operation) *Controller {
	return &Controller{
		vldtr:            vldtr,
		usecaseAuth:      usecaseAuth,
		usecaseOperation: usecaseOperation,
	}
}
