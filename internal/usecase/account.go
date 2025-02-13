package usecase

import (
	"context"
	"log/slog"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/internal/config"
	"github.com/m11ano/avito-shop/internal/domain"
)

type Account interface {
	GetItemByUsername(context.Context, string) (*domain.Account, error)
	Create(context.Context, *domain.Account) error
}

type AccountRepository interface {
	FindItemByUsername(context.Context, string) (*domain.Account, error)
	Create(context.Context, *domain.Account) error
	Update(context.Context, *domain.Account, uuid.UUID) error
}

type AccountInpl struct {
	logger    *slog.Logger
	config    config.Config
	repo      AccountRepository
	txManager *manager.Manager
}

func NewAccountInpl(logger *slog.Logger, config config.Config, txManager *manager.Manager, repo AccountRepository) *AccountInpl {
	uc := &AccountInpl{
		logger:    logger,
		config:    config,
		txManager: txManager,
		repo:      repo,
	}
	return uc
}

func (uc *AccountInpl) GetItemByUsername(ctx context.Context, username string) (*domain.Account, error) {
	return uc.repo.FindItemByUsername(ctx, username)
}

func (uc *AccountInpl) Create(ctx context.Context, item *domain.Account) error {
	return uc.repo.Create(ctx, item)
}
