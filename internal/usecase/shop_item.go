package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/internal/domain"
)

type ShopItem interface {
	GetItemByID(context.Context, uuid.UUID) (*domain.ShopItem, error)
	GetItemByName(context.Context, string) (*domain.ShopItem, error)
	GetItemsByIDs(context.Context, []uuid.UUID) ([]domain.ShopItem, error)
}

type ShopItemRepository interface {
	FindItemByID(context.Context, uuid.UUID) (*domain.ShopItem, error)
	FindItemByName(context.Context, string) (*domain.ShopItem, error)
	FindItemsByIDs(context.Context, []uuid.UUID) ([]domain.ShopItem, error)
}

type ShopItemInpl struct {
	repo ShopItemRepository
}

func NewShopItemInpl(repo ShopItemRepository) *ShopItemInpl {
	uc := &ShopItemInpl{
		repo: repo,
	}
	return uc
}

func (uc *ShopItemInpl) GetItemByID(ctx context.Context, id uuid.UUID) (*domain.ShopItem, error) {
	return uc.repo.FindItemByID(ctx, id)
}

func (uc *ShopItemInpl) GetItemByName(ctx context.Context, name string) (*domain.ShopItem, error) {
	return uc.repo.FindItemByName(ctx, name)
}

func (uc *ShopItemInpl) GetItemsByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.ShopItem, error) {
	return uc.repo.FindItemsByIDs(ctx, ids)
}
