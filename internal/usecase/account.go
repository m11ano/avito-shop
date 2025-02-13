package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/internal/app"
	"github.com/m11ano/avito-shop/internal/config"
	"github.com/m11ano/avito-shop/internal/domain"
)

const maxTxAttempts = 3

type Account interface {
	SignInOrSignUp(context.Context, string, string) (string, error)
	AuthByJWTToken(context.Context, string) (*uuid.UUID, error)
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

func (uc *AccountInpl) generateJWTToken(ctx context.Context, account *domain.Account) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"accountID": account.ID.String(),
		"createdAt": strconv.FormatInt(time.Now().Unix(), 10),
	})

	tokenStr, err := token.SignedString([]byte(uc.config.Account.JWTSecretKey))
	if err != nil {
		uc.logger.ErrorContext(ctx, "jwt sign error", slog.Any("error", err))
		return "", app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	return tokenStr, nil
}

func (uc *AccountInpl) SignInOrSignUp(ctx context.Context, username string, password string) (string, error) {
	var account *domain.Account
	var err error

	for i := 0; i < maxTxAttempts; i++ {
		err = uc.txManager.Do(ctx, func(ctx context.Context) error {
			account, err = uc.repo.FindItemByUsername(ctx, username)
			if err != nil && errors.Is(err, app.ErrNotFound) {
				account, err = domain.NewAccount(username, password)
				if err != nil {
					return err
				}

				err = uc.repo.Create(ctx, account)
				if err != nil {
					return err
				}
			} else if err != nil {
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
			return "", app.NewErrorFrom(app.ErrInternal).Wrap(err)
		}
		return "", err
	}

	if account == nil {
		return "", app.NewErrorFrom(app.ErrInternal)
	}

	check, err := account.VerifyPassword(password)
	if err != nil {
		return "", app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	if !check {
		return "", app.ErrUnauthorized
	}

	return uc.generateJWTToken(ctx, account)
}

func (uc *AccountInpl) AuthByJWTToken(ctx context.Context, tokenStr string) (*uuid.UUID, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(_ *jwt.Token) (interface{}, error) {
		return []byte(uc.config.Account.JWTSecretKey), nil
	})
	if err != nil {
		uc.logger.ErrorContext(ctx, "parse jwt", slog.Any("error", err))
		return nil, app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	if !token.Valid {
		return nil, app.ErrUnauthorized
	}

	var accountIDStr string
	var createdAtStr string
	var createdAt int64
	var ok bool

	if accountIDStr, ok = claims["accountID"].(string); !ok {
		return nil, app.ErrUnauthorized
	}

	if createdAtStr, ok = claims["createdAt"].(string); !ok {
		return nil, app.ErrUnauthorized
	}

	createdAt, err = strconv.ParseInt(createdAtStr, 10, 64)
	if err != nil {
		return nil, app.NewErrorFrom(app.ErrUnauthorized).Wrap(err)
	}

	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return nil, app.NewErrorFrom(app.ErrUnauthorized).Wrap(err)
	}

	if time.Now().Unix()-createdAt > uc.config.Account.JWTTokenTTL {
		return nil, app.ErrUnauthorized
	}

	return &accountID, nil
}
