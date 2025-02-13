package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/m11ano/avito-shop/internal/app"
	"github.com/m11ano/avito-shop/internal/usecase"
)

func Auth(accountUsecase usecase.Account) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		isAuth := false

		AuthorizationHeader := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")

		if len(AuthorizationHeader) > 0 {
			accountID, err := accountUsecase.AuthByJWTToken(c.Context(), AuthorizationHeader)
			if err != nil && !errors.Is(err, app.ErrUnauthorized) {
				return err
			}

			if accountID != nil {
				isAuth = true
				c.Locals("authAccountID", accountID)
			}
		}

		c.Locals("isAuth", isAuth)

		return c.Next()
	}
}
