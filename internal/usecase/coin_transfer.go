package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/internal/app"
	"github.com/m11ano/avito-shop/internal/config"
	"github.com/m11ano/avito-shop/internal/domain"
)

const coinTransferMaxTxAttempts = 3

type CoinTransferGetAggrHistoryItem struct {
	Account *domain.Account
	Amount  int64
}

type CoinTransfer interface {
	MakeTransferByUsername(ctx context.Context, targetAccountUsername string, ownerAccountID uuid.UUID, amount int64, identityKey *uuid.UUID) (*domain.CoinTransfer, *domain.CoinTransfer, error)
	GetAggrCoinHistory(ctx context.Context, accountID uuid.UUID, transferType domain.CoinTransferType) ([]CoinTransferGetAggrHistoryItem, error)
}

type CoinTransferRepositoryAggrHistoryItem struct {
	AccountID uuid.UUID
	Ammount   int64
}

type CoinTransferRepository interface {
	FindIdentity(context.Context, uuid.UUID) (bool, error)
	Create(context.Context, *domain.CoinTransfer) error
	GetAggrCoinHistoryByAccountID(context.Context, uuid.UUID, domain.CoinTransferType) ([]CoinTransferRepositoryAggrHistoryItem, error)
}

type CoinTransferInpl struct {
	logger           *slog.Logger
	config           config.Config
	repo             CoinTransferRepository
	txManager        *manager.Manager
	usecaseAccount   Account
	usecaseOperation Operation
}

func NewCoinTransferInpl(logger *slog.Logger, config config.Config, txManager *manager.Manager, repo CoinTransferRepository, usecaseAccount Account, usecaseOperation Operation) *CoinTransferInpl {
	uc := &CoinTransferInpl{
		logger:           logger,
		config:           config,
		txManager:        txManager,
		repo:             repo,
		usecaseAccount:   usecaseAccount,
		usecaseOperation: usecaseOperation,
	}
	return uc
}

// Make coin transfer from owner to target
// First coint transfer is for owner, second for target
func (uc *CoinTransferInpl) MakeTransferByUsername(ctx context.Context, targetAccountUsername string, ownerAccountID uuid.UUID, ammount int64, identityKey *uuid.UUID) (*domain.CoinTransfer, *domain.CoinTransfer, error) {
	var err error
	var transferForOwner *domain.CoinTransfer
	var transferForTarget *domain.CoinTransfer

	for i := 0; i < coinTransferMaxTxAttempts; i++ {
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

			targetAccount, err := uc.usecaseAccount.GetItemByUsername(ctx, targetAccountUsername)
			if err != nil {
				return err
			}

			if targetAccount.ID == ownerAccountID {
				return app.NewErrorFrom(app.ErrConflict).SetMessage("cant send coin to yourself")
			}

			transferForTarget = domain.NewCoinTransfer(domain.CoinTransferTypeReciving, ownerAccountID, targetAccount.ID, ammount, identityKey)
			transferForOwner = domain.NewCoinTransfer(domain.CoinTransferTypeSending, targetAccount.ID, ownerAccountID, ammount, identityKey)

			operationForTarget := domain.NewOperation(domain.OperationTypeIncrease, targetAccount.ID, ammount, domain.OperationSourceTypeTransfer, &transferForTarget.ID)
			operationForOwner := domain.NewOperation(domain.OperationTypeDecrease, ownerAccountID, ammount, domain.OperationSourceTypeTransfer, &transferForOwner.ID)

			err = uc.usecaseOperation.SaveOperation(ctx, operationForOwner)
			if err != nil {
				return err
			}

			err = uc.usecaseOperation.SaveOperation(ctx, operationForTarget)
			if err != nil {
				return err
			}

			err = uc.repo.Create(ctx, transferForOwner)
			if err != nil {
				return err
			}

			err = uc.repo.Create(ctx, transferForTarget)
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
			return nil, nil, app.NewErrorFrom(app.ErrInternal).Wrap(err)
		}
		return nil, nil, err
	}

	return transferForOwner, transferForTarget, nil
}

func (uc *CoinTransferInpl) GetAggrCoinHistory(ctx context.Context, accountID uuid.UUID, transferType domain.CoinTransferType) ([]CoinTransferGetAggrHistoryItem, error) {
	result := make([]CoinTransferGetAggrHistoryItem, 0)

	history, err := uc.repo.GetAggrCoinHistoryByAccountID(ctx, accountID, transferType)
	if err != nil {
		return nil, err
	}
	accountIDs := make([]uuid.UUID, 0, len(history))
	for _, item := range history {
		accountIDs = append(accountIDs, item.AccountID)
	}

	accountItems, err := uc.usecaseAccount.GetItemsByIDs(ctx, accountIDs)
	if err != nil {
		return nil, err
	}

	for _, invItem := range history {
		var hstrItem *domain.Account
		for _, accountItem := range accountItems {
			if accountItem.ID == invItem.AccountID {
				hstrItem = &accountItem
				break
			}
		}
		result = append(result, CoinTransferGetAggrHistoryItem{
			Account: hstrItem,
			Amount:  invItem.Ammount,
		})
	}

	return result, nil
}
