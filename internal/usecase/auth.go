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

const authMaxTxAttempts = 3

type Auth interface {
	SignInOrSignUp(context.Context, string, string) (string, error)
	AuthByJWTToken(context.Context, string) (*uuid.UUID, error)
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

func (uc *AuthInpl) generateJWTToken(ctx context.Context, auth *domain.Account) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authID":    auth.ID.String(),
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

	for i := 0; i < authMaxTxAttempts; i++ {
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
					err = uc.usecaseOperation.AddOperation(ctx, depositOp)
					if err != nil {
						return err
					}
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

func (uc *AuthInpl) AuthByJWTToken(ctx context.Context, tokenStr string) (*uuid.UUID, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(_ *jwt.Token) (interface{}, error) {
		return []byte(uc.config.Auth.JWTSecretKey), nil
	})
	if err != nil {
		uc.logger.ErrorContext(ctx, "parse jwt", slog.Any("error", err))
		return nil, app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	if !token.Valid {
		return nil, app.ErrUnauthorized
	}

	var authIDStr string
	var createdAtStr string
	var createdAt int64
	var ok bool

	if authIDStr, ok = claims["authID"].(string); !ok {
		return nil, app.ErrUnauthorized
	}

	if createdAtStr, ok = claims["createdAt"].(string); !ok {
		return nil, app.ErrUnauthorized
	}

	createdAt, err = strconv.ParseInt(createdAtStr, 10, 64)
	if err != nil {
		return nil, app.NewErrorFrom(app.ErrUnauthorized).Wrap(err)
	}

	authID, err := uuid.Parse(authIDStr)
	if err != nil {
		return nil, app.NewErrorFrom(app.ErrUnauthorized).Wrap(err)
	}

	if time.Now().Unix()-createdAt > uc.config.Auth.JWTTokenTTL {
		return nil, app.ErrUnauthorized
	}

	return &authID, nil
}
