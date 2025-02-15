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

var ErrOperationNotEnoughFunds = app.NewErrorFrom(app.ErrUnprocessableEntity).SetMessage("not enough funds")

type Operation interface {
	GetBalanceByAccountID(ctx context.Context, accountID uuid.UUID) (balance int64, found bool, err error)
	SaveOperation(ctx context.Context, operation *domain.Operation) (balance int64, err error)
}

type OperationRepository interface {
	GetBalanceByAccountID(ctx context.Context, accountID uuid.UUID) (balance int64, found bool, err error)
	Create(ctx context.Context, operation *domain.Operation) (balance int64, err error)
}

type OperationInpl struct {
	logger    *slog.Logger
	config    config.Config
	repo      OperationRepository
	txManager *manager.Manager
}

func NewOperationInpl(logger *slog.Logger, config config.Config, txManager *manager.Manager, repo OperationRepository) *OperationInpl {
	uc := &OperationInpl{
		logger:    logger,
		config:    config,
		txManager: txManager,
		repo:      repo,
	}
	return uc
}

func (uc *OperationInpl) SaveOperation(ctx context.Context, operation *domain.Operation) (int64, error) {
	var err error
	var balance int64
	switch operation.Type {
	case domain.OperationTypeDecrease:
		// Если списание - проверим баланс в транзакции и если меньше 0 - откатим принудительно и вернем ошибку
		err = uc.txManager.Do(ctx, func(ctx context.Context) error {
			balance, err = uc.repo.Create(ctx, operation)
			if err != nil {
				return err
			}

			if balance < 0 {
				return ErrOperationNotEnoughFunds
			}

			return nil
		})
		if err != nil {
			return 0, err
		}
	case domain.OperationTypeIncrease:
		// Если пополнение - просто сохраняем
		balance, err = uc.repo.Create(ctx, operation)
		if err != nil {
			return 0, err
		}
	}

	return balance, nil
}

func (uc *OperationInpl) GetBalanceByAccountID(ctx context.Context, accountID uuid.UUID) (int64, bool, error) {
	return uc.repo.GetBalanceByAccountID(ctx, accountID)
}
