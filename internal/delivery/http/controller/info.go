package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/internal/domain"
	"github.com/m11ano/avito-shop/pkg/e"
)

type InfoHandlerOut struct {
	Coins       int64                     `json:"coins"`
	Inventory   []InfoHandlerOutInventory `json:"inventory"`
	CoinHistory InfoHandlerOutCoinHistory `json:"coinHistory"`
}

type InfoHandlerOutInventory struct {
	Type     string `json:"type"`
	Quantity int64  `json:"quantity"`
}

type InfoHandlerOutCoinHistory struct {
	Received []InfoHandlerOutCoinHistoryReceived `json:"received"`
	Sent     []InfoHandlerOutCoinHistorySent     `json:"sent"`
}

type InfoHandlerOutCoinHistorySent struct {
	ToUser string `json:"toUser"`
	Amount int64  `json:"amount"`
}

type InfoHandlerOutCoinHistoryReceived struct {
	FromUser string `json:"fromUser"`
	Amount   int64  `json:"amount"`
}

func (ctrl *Controller) InfoHandler(c *fiber.Ctx) error {
	if isAuth, ok := c.Locals("isAuth").(bool); !ok || !isAuth {
		return e.ErrUnauthorized
	}

	var accountID *uuid.UUID
	var ok bool
	if accountID, ok = c.Locals("authAccountID").(*uuid.UUID); !ok {
		return e.ErrUnauthorized
	}

	var err error
	out := InfoHandlerOut{}

	// Получаем баланс пользователя
	out.Coins, _, err = ctrl.usecaseOperation.GetBalanceByAccountID(c.Context(), *accountID)
	if err != nil {
		return err
	}

	// Получаем инвентарь и перекладываем в дто ответа
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

	// Получаем агрегированную историю полученных монет и перекладываем в дто ответа
	receivedCoinHistory, err := ctrl.usecaseCoinTransfer.GetAggrCoinHistory(c.Context(), *accountID, domain.CoinTransferTypeReciving)
	if err != nil {
		return err
	}

	out.CoinHistory.Received = make([]InfoHandlerOutCoinHistoryReceived, 0, len(receivedCoinHistory))
	for _, item := range receivedCoinHistory {
		out.CoinHistory.Received = append(out.CoinHistory.Received, InfoHandlerOutCoinHistoryReceived{
			FromUser: item.Account.Username,
			Amount:   item.Amount,
		})
	}

	// Получаем агрегированную историю отправленных монет и перекладываем в дто ответа
	sentCoinHistory, err := ctrl.usecaseCoinTransfer.GetAggrCoinHistory(c.Context(), *accountID, domain.CoinTransferTypeSending)
	if err != nil {
		return err
	}

	out.CoinHistory.Sent = make([]InfoHandlerOutCoinHistorySent, 0, len(sentCoinHistory))
	for _, item := range sentCoinHistory {
		out.CoinHistory.Sent = append(out.CoinHistory.Sent, InfoHandlerOutCoinHistorySent{
			ToUser: item.Account.Username,
			Amount: item.Amount,
		})
	}

	return c.JSON(out)
}
