package usecase

import (
	"context"
	"log/slog"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/internal/app"
	"github.com/m11ano/avito-shop/internal/config"
	"github.com/m11ano/avito-shop/internal/domain"
)

type CoinTransferGetAggrHistoryItem struct {
	Account *domain.Account
	Amount  int64
}

//go:generate mockery --name=CoinTransfer --output=../../tests/mocks --case=underscore
type CoinTransfer interface {
	MakeTransferByUsername(ctx context.Context, targetAccountUsername string, ownerAccountID uuid.UUID, amount int64, identityKey *uuid.UUID) (ownerCoinTransfer *domain.CoinTransfer, targetCoinTransfer *domain.CoinTransfer, err error)
	GetAggrCoinHistory(ctx context.Context, accountID uuid.UUID, transferType domain.CoinTransferType) (aggrHistory []CoinTransferGetAggrHistoryItem, err error)
}

type CoinTransferRepositoryAggrHistoryItem struct {
	AccountID uuid.UUID
	Ammount   int64
}

//go:generate mockery --name=CoinTransferRepository --output=../../tests/mocks --case=underscore
type CoinTransferRepository interface {
	FindIdentity(ctx context.Context, identityKey uuid.UUID) (found bool, err error)
	Create(ctx context.Context, coinTransfer *domain.CoinTransfer) error
	GetAggrCoinHistoryByAccountID(ctx context.Context, accountID uuid.UUID, transferType domain.CoinTransferType) (aggrHistory []CoinTransferRepositoryAggrHistoryItem, err error)
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
func (uc *CoinTransferInpl) MakeTransferByUsername(ctx context.Context, targetAccountUsername string, ownerAccountID uuid.UUID, ammount int64, identityKey *uuid.UUID) (*domain.CoinTransfer, *domain.CoinTransfer, error) {
	var err error
	var transferForOwner *domain.CoinTransfer
	var transferForTarget *domain.CoinTransfer

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

		_, err = uc.usecaseOperation.SaveOperation(ctx, operationForOwner)
		if err != nil {
			return err
		}

		_, err = uc.usecaseOperation.SaveOperation(ctx, operationForTarget)
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
	if err != nil {
		if !app.IsAppError(err) {
			return nil, nil, app.NewErrorFrom(app.ErrInternal).Wrap(err)
		}
		return nil, nil, err
	}

	return transferForOwner, transferForTarget, nil
}

func (uc *CoinTransferInpl) GetAggrCoinHistory(ctx context.Context, accountID uuid.UUID, transferType domain.CoinTransferType) ([]CoinTransferGetAggrHistoryItem, error) {
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

	result := make([]CoinTransferGetAggrHistoryItem, 0, len(history))

	for _, item := range history {
		resultItem := CoinTransferGetAggrHistoryItem{
			Amount: item.Ammount,
		}
		if accountItem, ok := accountItems[item.AccountID]; ok {
			resultItem.Account = &accountItem
		}

		result = append(result, resultItem)
	}

	return result, nil
}
