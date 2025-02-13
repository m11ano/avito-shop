package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/m11ano/avito-shop/internal/config"
	"github.com/m11ano/avito-shop/internal/delivery/http/handler"
)

func RegisterRoutes(app *fiber.App, config config.Config) {
	rootGroup := app.Group(config.HTTP.Prefix)

	h := handler.New()

	rootGroup.Get("/test", h.Test)
}
