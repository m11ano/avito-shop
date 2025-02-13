package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/internal/app"
)

type InfoHandlerOut struct {
	Coins     int64                     `json:"coins"`
	Inventory []InfoHandlerOutInventory `json:"inventory"`
}

type InfoHandlerOutInventory struct {
	Type     string `json:"type"`
	Quantity int64  `json:"quantity"`
}

func (ctrl *Controller) InfoHandler(c *fiber.Ctx) error {
	if isAuth, ok := c.Locals("isAuth").(bool); !ok || !isAuth {
		return app.ErrUnauthorized
	}

	var accountID *uuid.UUID
	var ok bool
	if accountID, ok = c.Locals("authAccountID").(*uuid.UUID); !ok {
		return app.ErrUnauthorized
	}

	var err error
	out := InfoHandlerOut{}

	out.Coins, _, err = ctrl.usecaseOperation.GetBalanceByAccountID(c.Context(), *accountID)
	if err != nil {
		return err
	}

	inventory, err := ctrl.usecaseShopPurchase.GetInventory(c.Context(), *accountID)
	if err != nil {
		return err
	}

	out.Inventory = make([]InfoHandlerOutInventory, 0, len(inventory))
	for _, item := range inventory {
		invItem := InfoHandlerOutInventory{
			Quantity: item.Quantity,
		}
		if item.ShopItem != nil {
			invItem.Type = item.ShopItem.Name
		}
		out.Inventory = append(out.Inventory, invItem)
	}

	return c.JSON(out)
}
