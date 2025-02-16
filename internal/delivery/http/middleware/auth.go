package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/m11ano/avito-shop/internal/usecase"
	"github.com/m11ano/avito-shop/pkg/e"
)

func Auth(authUsecase usecase.Auth) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		isAuth := false

		AuthorizationHeader := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")

		if len(AuthorizationHeader) > 0 {
			accountID, err := authUsecase.AuthByJWTToken(c.Context(), AuthorizationHeader)
			if err != nil && !errors.Is(err, e.ErrUnauthorized) {
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
