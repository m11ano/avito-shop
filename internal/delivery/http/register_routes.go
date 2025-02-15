package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/m11ano/avito-shop/internal/config"
	"github.com/m11ano/avito-shop/internal/delivery/http/controller"
	"github.com/m11ano/avito-shop/internal/delivery/http/middleware"
	"github.com/m11ano/avito-shop/internal/usecase"
)

func RegisterRoutes(app *fiber.App, config config.Config, ctrl *controller.Controller, authUsecase usecase.Auth) {
	authMiddleware := middleware.Auth(authUsecase)
	rootGroup := app.Group(config.HTTP.Prefix, authMiddleware)

	rootGroup.Post("/auth", ctrl.AuthHandler)
	rootGroup.Get("/info", ctrl.InfoHandler)
	rootGroup.Get("/buy/:name", ctrl.BuyHandler)
	rootGroup.Post("/sendCoin", ctrl.SendCoinHandler)
}
