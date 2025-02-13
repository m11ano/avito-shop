package handler

import "github.com/gofiber/fiber/v2"

func (h *Handler) Test(c *fiber.Ctx) error {
	return c.SendString("its works")
}
