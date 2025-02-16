package controller

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/m11ano/avito-shop/internal/delivery/http/validation"
	"github.com/m11ano/avito-shop/pkg/e"
)

type AuthHandlerOut struct {
	Token string `json:"token"`
}

type AuthHandlerIn struct {
	Username string `json:"username" validate:"required,max=255"`
	Password string `json:"password" validate:"required,max=255"`
}

func (ctrl *Controller) AuthHandlerValidate(in *AuthHandlerIn) (isOk bool, errMsg string) {
	if err := ctrl.vldtr.Struct(in); err != nil {
		return validation.FormatErrors(err)
	}
	return true, ""
}

func (ctrl *Controller) AuthHandler(c *fiber.Ctx) error {
	in := &AuthHandlerIn{}

	if err := c.BodyParser(in); err != nil {
		return &fiber.Error{Code: fiber.ErrBadRequest.Code, Message: err.Error()}
	}

	ok, errMsg := ctrl.AuthHandlerValidate(in)
	if !ok {
		return &fiber.Error{Code: fiber.ErrBadRequest.Code, Message: errMsg}
	}

	jwtToken, err := ctrl.usecaseAuth.SignInOrSignUp(c.Context(), in.Username, in.Password)
	if err != nil {
		if errors.Is(err, e.ErrInternal) {
			return err
		}
		return e.ErrUnauthorized
	}

	out := AuthHandlerOut{Token: jwtToken}

	return c.JSON(out)
}
