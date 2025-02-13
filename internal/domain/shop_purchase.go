package domain

import (
	"time"

	"github.com/google/uuid"
)

type ShopPurchase struct {
	ID          uuid.UUID
	ItemID      uuid.UUID
	AccountID   uuid.UUID
	Quantity    int64
	CreatedAt   time.Time
	IdentityKey *uuid.UUID
}

func NewShopPurchase(itemID uuid.UUID, accountID uuid.UUID, quantity int64, identityKey *uuid.UUID) *ShopPurchase {
	return &ShopPurchase{
		ID:          uuid.New(),
		ItemID:      itemID,
		AccountID:   accountID,
		Quantity:    quantity,
		CreatedAt:   time.Now(),
		IdentityKey: identityKey,
	}
}
