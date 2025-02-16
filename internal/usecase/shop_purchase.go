package usecase

import (
	"context"
	"log/slog"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/internal/domain"
	"github.com/m11ano/avito-shop/internal/infra/config"
	"github.com/m11ano/avito-shop/pkg/e"
)

type ShopPurchaseGetInventoryItem struct {
	ShopItem *domain.ShopItem
	Quantity int64
}

//go:generate mockery --name=ShopPurchase --output=../../tests/mocks --case=underscore
type ShopPurchase interface {
	MakePurchase(ctx context.Context, shopItemName string, ownerAccountID uuid.UUID, quantity int64, identityKey *uuid.UUID) (shopPurchase *domain.ShopPurchase, err error)
	GetInventory(ctx context.Context, accountID uuid.UUID) (inventory []ShopPurchaseGetInventoryItem, err error)
}

type ShopPurchaseRepositoryAggrInventoryItem struct {
	ShopItemID uuid.UUID
	Quantity   int64
}

//go:generate mockery --name=ShopPurchaseRepository --output=../../tests/mocks --case=underscore
type ShopPurchaseRepository interface {
	FindIdentity(ctx context.Context, identityKey uuid.UUID) (found bool, err error)
	Create(ctx context.Context, shopPurchase *domain.ShopPurchase) error
	AggrInventoryByAccountID(ctx context.Context, accountID uuid.UUID) (inventory []ShopPurchaseRepositoryAggrInventoryItem, err error)
}

type ShopPurchaseInpl struct {
	logger           *slog.Logger
	config           config.Config
	repo             ShopPurchaseRepository
	txManager        *manager.Manager
	usecaseAccount   Account
	usecaseOperation Operation
	usecaseShopItem  ShopItem
}

func NewShopPurchaseInpl(logger *slog.Logger, config config.Config, txManager *manager.Manager, repo ShopPurchaseRepository, usecaseAccount Account, usecaseOperation Operation, usecaseShopItem ShopItem) *ShopPurchaseInpl {
	uc := &ShopPurchaseInpl{
		logger:           logger,
		config:           config,
		txManager:        txManager,
		repo:             repo,
		usecaseAccount:   usecaseAccount,
		usecaseOperation: usecaseOperation,
		usecaseShopItem:  usecaseShopItem,
	}
	return uc
}

func (uc *ShopPurchaseInpl) MakePurchase(ctx context.Context, itemName string, accountID uuid.UUID, quantity int64, identityKey *uuid.UUID) (*domain.ShopPurchase, error) {
	var err error
	var shopPurchase *domain.ShopPurchase

	err = uc.txManager.Do(ctx, func(ctx context.Context) error {
		if identityKey != nil {
			isIdentityExists, err := uc.repo.FindIdentity(ctx, *identityKey)
			if err != nil {
				return err
			}

			if isIdentityExists {
				return e.ErrConflict
			}
		}

		shopItem, err := uc.usecaseShopItem.GetItemByName(ctx, itemName)
		if err != nil {
			return err
		}

		shopPurchase = domain.NewShopPurchase(shopItem.ID, accountID, quantity, identityKey)

		operation := domain.NewOperation(domain.OperationTypeDecrease, accountID, shopItem.Price*quantity, domain.OperationSourceTypeShopPurchase, &shopPurchase.ID)

		_, err = uc.usecaseOperation.SaveOperation(ctx, operation)
		if err != nil {
			return err
		}

		err = uc.repo.Create(ctx, shopPurchase)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if !e.IsAppError(err) {
			return nil, e.NewErrorFrom(e.ErrInternal).Wrap(err)
		}
		return nil, err
	}

	return shopPurchase, nil
}

func (uc *ShopPurchaseInpl) GetInventory(ctx context.Context, accountID uuid.UUID) ([]ShopPurchaseGetInventoryItem, error) {
	inventory, err := uc.repo.AggrInventoryByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	shopItemsIDs := make([]uuid.UUID, 0, len(inventory))
	for _, item := range inventory {
		shopItemsIDs = append(shopItemsIDs, item.ShopItemID)
	}

	shopItems, err := uc.usecaseShopItem.GetItemsByIDs(ctx, shopItemsIDs)
	if err != nil {
		return nil, err
	}

	result := make([]ShopPurchaseGetInventoryItem, 0, len(inventory))

	for _, item := range inventory {
		resultItem := ShopPurchaseGetInventoryItem{
			Quantity: item.Quantity,
		}
		if shopItem, ok := shopItems[item.ShopItemID]; ok {
			resultItem.ShopItem = &shopItem
		}

		result = append(result, resultItem)
	}

	return result, nil
}
