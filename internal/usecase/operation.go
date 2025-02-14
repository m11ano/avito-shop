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

const operationMaxTxAttempts = 3

var ErrOperationNotEnoughFunds = app.NewErrorFrom(app.ErrUnprocessableEntity).SetMessage("not enough funds")

type Operation interface {
	GetBalanceByAccountID(context.Context, uuid.UUID) (int64, int64, error)
	SaveOperation(context.Context, *domain.Operation) error
}

type OperationRepository interface {
	CountBalanceByAccountID(context.Context, uuid.UUID) (int64, int64, error)
	Create(context.Context, *domain.Operation) error
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

func (uc *OperationInpl) SaveOperation(ctx context.Context, operation *domain.Operation) error {
	switch operation.Type {
	case domain.OperationTypeDecrease:
		// Если списание - проверим предварительно баланс в транзакции
		var err error
		for i := 0; i < operationMaxTxAttempts; i++ {
			err = uc.txManager.Do(ctx, func(ctx context.Context) error {
				balance, _, err := uc.repo.CountBalanceByAccountID(ctx, operation.AccountID)
				if err != nil {
					return err
				}

				if balance-operation.Amount < 0 {
					return ErrOperationNotEnoughFunds
				}

				err = uc.repo.Create(ctx, operation)
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
				return app.NewErrorFrom(app.ErrInternal).Wrap(err)
			}
			return err
		}
	case domain.OperationTypeIncrease:
		// Если пополнение - просто сохраняем
		err := uc.repo.Create(ctx, operation)
		if err != nil {
			return err
		}
	}

	return nil
}

func (uc *OperationInpl) GetBalanceByAccountID(ctx context.Context, accountID uuid.UUID) (int64, int64, error) {
	return uc.repo.CountBalanceByAccountID(ctx, accountID)
}
