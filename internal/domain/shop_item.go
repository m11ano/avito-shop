package domain

import "github.com/google/uuid"

type ShopItem struct {
	ID    uuid.UUID
	Name  string
	Price int64
}

func NewShopItem(name string, price int64) *ShopItem {
	return &ShopItem{
		ID:    uuid.New(),
		Name:  name,
		Price: price,
	}
}
