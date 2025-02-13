package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/internal/app"
	"github.com/m11ano/avito-shop/internal/config"
	"github.com/m11ano/avito-shop/internal/domain"
)

const shopPurchaseMaxTxAttempts = 3

type ShopPurchaseGetInventoryItem struct {
	ShopItem *domain.ShopItem
	Quantity int64
}

type ShopPurchase interface {
	MakePurchase(context.Context, string, uuid.UUID, int64, *uuid.UUID) (*domain.ShopPurchase, error)
	GetInventory(context.Context, uuid.UUID) ([]ShopPurchaseGetInventoryItem, error)
}

type ShopPurchaseRepositoryAggrInventoryItem struct {
	ShopItemID uuid.UUID
	Quantity   int64
}

type ShopPurchaseRepository interface {
	FindIdentity(context.Context, uuid.UUID) (bool, error)
	Create(context.Context, *domain.ShopPurchase) error
	AggrInventoryByAccountID(context.Context, uuid.UUID) ([]ShopPurchaseRepositoryAggrInventoryItem, error)
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

	for i := 0; i < shopPurchaseMaxTxAttempts; i++ {
		err = uc.txManager.Do(ctx, func(ctx context.Context) error {
			if identityKey != nil {
				isIdentityExists, err := uc.repo.FindIdentity(ctx, *identityKey)
				if err != nil {
					return err
				}

				if isIdentityExists {
					return app.ErrConflict
				}
			}

			item, err := uc.usecaseShopItem.GetItemByName(ctx, itemName)
			if err != nil {
				return err
			}

			// Если бы был учет количества - здесь нужно было бы это проверить =)

			shopPurchase = domain.NewShopPurchase(item.ID, accountID, quantity, identityKey)

			operation := domain.NewOperation(domain.OperationTypeDecrease, accountID, item.Price*quantity, domain.OperationSourceTypeShopPurchase, &shopPurchase.ID)

			err = uc.usecaseOperation.AddOperation(ctx, operation)
			if err != nil {
				return err
			}

			err = uc.repo.Create(ctx, shopPurchase)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil && app.ErrCheckIsTxСoncurrentExec(err) {
			time.Sleep(time.Duration((i+1)*100) * time.Millisecond)
			continue
		}

		break
	}
	if err != nil {
		if !app.IsAppError(err) {
			return nil, app.NewErrorFrom(app.ErrInternal).Wrap(err)
		}
		return nil, err
	}

	return shopPurchase, nil
}

func (uc *ShopPurchaseInpl) GetInventory(ctx context.Context, accountID uuid.UUID) ([]ShopPurchaseGetInventoryItem, error) {
	result := make([]ShopPurchaseGetInventoryItem, 0)

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

	for _, invItem := range inventory {
		var invShopItem *domain.ShopItem
		for _, shopItem := range shopItems {
			if shopItem.ID == invItem.ShopItemID {
				invShopItem = &shopItem
				break
			}
		}
		result = append(result, ShopPurchaseGetInventoryItem{
			ShopItem: invShopItem,
			Quantity: invItem.Quantity,
		})
	}

	fmt.Println(inventory)
	fmt.Println(shopItems)

	return result, nil
}
