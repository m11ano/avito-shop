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

//go:generate mockery --name=Auth --output=../../tests/mocks --case=underscore
type Auth interface {
	SignInOrSignUp(ctx context.Context, username string, password string) (jwtToken string, err error)
	AuthByJWTToken(ctx context.Context, jwtToken string) (accountID *uuid.UUID, err error)
}

type AuthInpl struct {
	logger           *slog.Logger
	config           config.Config
	txManager        *manager.Manager
	usecaseAccount   Account
	usecaseOperation Operation
}

func NewAuthInpl(logger *slog.Logger, config config.Config, txManager *manager.Manager, usecaseAccount Account, usecaseOperation Operation) *AuthInpl {
	uc := &AuthInpl{
		logger:           logger,
		config:           config,
		txManager:        txManager,
		usecaseAccount:   usecaseAccount,
		usecaseOperation: usecaseOperation,
	}
	return uc
}

func (uc *AuthInpl) generateJWTToken(ctx context.Context, account *domain.Account) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"accountID": account.ID.String(),
		"createdAt": strconv.FormatInt(time.Now().Unix(), 10),
	})

	tokenStr, err := token.SignedString([]byte(uc.config.Auth.JWTSecretKey))
	if err != nil {
		uc.logger.ErrorContext(ctx, "jwt sign error", slog.Any("error", err))
		return "", app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	return tokenStr, nil
}

func (uc *AuthInpl) SignInOrSignUp(ctx context.Context, username string, password string) (string, error) {
	var account *domain.Account
	var err error

	err = uc.txManager.Do(ctx, func(ctx context.Context) error {
		account, err = uc.usecaseAccount.GetItemByUsername(ctx, username)
		if err != nil && errors.Is(err, app.ErrNotFound) {
			account, err = domain.NewAccount(username, password)
			if err != nil {
				return err
			}

			err = uc.usecaseAccount.Create(ctx, account)
			if err != nil {
				return err
			}

			if uc.config.Auth.NewAccountAmount > 0 {
				depositOp := domain.NewOperation(domain.OperationTypeIncrease, account.ID, uc.config.Auth.NewAccountAmount, domain.OperationSourceTypeDeposit, nil)
				_, err := uc.usecaseOperation.SaveOperation(ctx, depositOp)
				if err != nil {
					return err
				}
			}
		} else if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if !app.IsAppError(err) {
			return "", app.NewErrorFrom(app.ErrInternal).Wrap(err)
		}
		return "", err
	}

	check := account.VerifyPassword(password)
	if !check {
		return "", app.ErrUnauthorized
	}

	return uc.generateJWTToken(ctx, account)
}

func (uc *AuthInpl) AuthByJWTToken(ctx context.Context, tokenStr string) (*uuid.UUID, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(_ *jwt.Token) (interface{}, error) {
		return []byte(uc.config.Auth.JWTSecretKey), nil
	})
	if err != nil {
		uc.logger.ErrorContext(ctx, "parse jwt", slog.Any("error", err))
		return nil, app.NewErrorFrom(app.ErrUnauthorized).Wrap(err)
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

	if time.Now().Unix()-createdAt > uc.config.Auth.JWTTokenTTL {
		return nil, app.ErrUnauthorized
	}

	return &accountID, nil
}
