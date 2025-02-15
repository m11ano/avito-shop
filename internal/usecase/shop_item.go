package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/internal/domain"
)

type ShopItem interface {
	GetItemByID(ctx context.Context, id uuid.UUID) (shopItem *domain.ShopItem, err error)
	GetItemByName(ctx context.Context, name string) (shopItem *domain.ShopItem, err error)
	GetItemsByIDs(ctx context.Context, ids []uuid.UUID) (shopItems []domain.ShopItem, err error)
}

type ShopItemRepository interface {
	FindItemByID(ctx context.Context, id uuid.UUID) (shopItem *domain.ShopItem, err error)
	FindItemByName(ctx context.Context, name string) (shopItem *domain.ShopItem, err error)
	FindItemsByIDs(ctx context.Context, ids []uuid.UUID) (shopItems []domain.ShopItem, err error)
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
