package domain

import "github.com/google/uuid"

type ShopItem struct {
	ID    uuid.UUID
	Name  string
	Price int64
}
