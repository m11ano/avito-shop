package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/Masterminds/squirrel"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m11ano/avito-shop/internal/app"
	"github.com/m11ano/avito-shop/internal/domain"
	"github.com/m11ano/avito-shop/internal/usecase"
	"github.com/m11ano/avito-shop/pkg/dbhelper"
)

const (
	coinTransferTable = "coin_transfer"
)

type DBCoinTransfer struct {
	ID                    uuid.UUID               `db:"transfer_id"`
	Type                  domain.CoinTransferType `db:"transfer_type"`
	OwnerAccountID        uuid.UUID               `db:"owner_account_id"`
	CounterpartyAccountID uuid.UUID               `db:"counterparty_account_id"`
	Amount                int64                   `db:"amount"`
	CreatedAt             time.Time               `db:"created_at"`
	IdentityKey           *uuid.UUID              `db:"identity_key"`
}

var (
	//nolint:unused
	coinTransferTableFields = []string{}
	coinTransferDBSchema    = &DBCoinTransfer{}
)

func init() {
	coinTransferTableFields = dbhelper.ExtractDBFields(coinTransferDBSchema)
}

type CoinTransfer struct {
	logger *slog.Logger
	db     *pgxpool.Pool
	txc    *trmpgx.CtxGetter
	qb     squirrel.StatementBuilderType
}

func NewCoinTransfer(logger *slog.Logger, db *pgxpool.Pool, txc *trmpgx.CtxGetter) *CoinTransfer {
	return &CoinTransfer{
		logger: logger,
		db:     db,
		txc:    txc,
		qb:     squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

//nolint:unused
func (r *CoinTransfer) dbToDomain(db *DBCoinTransfer) *domain.CoinTransfer {
	return &domain.CoinTransfer{
		ID:                    db.ID,
		Type:                  db.Type,
		OwnerAccountID:        db.OwnerAccountID,
		CounterpartyAccountID: db.CounterpartyAccountID,
		Amount:                db.Amount,
		CreatedAt:             db.CreatedAt,
		IdentityKey:           db.IdentityKey,
	}
}

type CoinTransferCheckIdentityDTO struct {
	Count int `db:"count"`
}

func (r *CoinTransfer) FindIdentity(ctx context.Context, key uuid.UUID) (bool, error) {
	query, args, err := r.qb.Select("COUNT(*) as count").From(coinTransferTable).Where(squirrel.Eq{"identity_key": key}).Limit(1).ToSql()
	if err != nil {
		r.logger.ErrorContext(ctx, "building query", slog.Any("error", err))
		return false, app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	rows, err := r.txc.DefaultTrOrDB(ctx, r.db).Query(ctx, query, args...)
	if err != nil {
		errIsConv, convErr := app.ErrConvertPgxToLogic(err)
		if !errIsConv {
			r.logger.ErrorContext(ctx, "executing query", slog.Any("error", err))
		}
		return false, convErr
	}

	defer rows.Close()

	dbData := &CoinTransferCheckIdentityDTO{}

	if err := pgxscan.ScanOne(dbData, rows); err != nil {
		errIsConv, convErr := app.ErrConvertPgxToLogic(err)
		if !errIsConv {
			r.logger.ErrorContext(ctx, "scan row", slog.Any("error", err))
		}
		return false, convErr
	}

	return dbData.Count > 0, nil
}

func (r *CoinTransfer) Create(ctx context.Context, item *domain.CoinTransfer) error {
	dataMap, err := dbhelper.StructToDBMap(item, coinTransferDBSchema)
	if err != nil {
		r.logger.ErrorContext(ctx, "convert struct to db map", slog.Any("error", err))
		return app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	query, args, err := r.qb.Insert(coinTransferTable).SetMap(dataMap).ToSql()
	if err != nil {
		r.logger.ErrorContext(ctx, "building query", slog.Any("error", err))
		return app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	_, err = r.txc.DefaultTrOrDB(ctx, r.db).Exec(ctx, query, args...)
	if err != nil {
		errIsConv, convErr := app.ErrConvertPgxToLogic(err)
		if !errIsConv {
			r.logger.ErrorContext(ctx, "executing query", slog.Any("error", err))
		}
		return convErr
	}

	return nil
}

type CoinTransferAggrCoinHistoryItem struct {
	AccountID   uuid.UUID `db:"counterparty_account_id"`
	TotalAmount int64     `db:"total_amount"`
}

func (r *CoinTransfer) GetAggrCoinHistoryByAccountID(ctx context.Context, accountID uuid.UUID, transferType domain.CoinTransferType) ([]usecase.CoinTransferRepositoryAggrHistoryItem, error) {
	query, args, err := r.qb.Select("counterparty_account_id", "SUM(amount) as total_amount").From(coinTransferTable).Where(squirrel.Eq{"owner_account_id": accountID, "transfer_type": transferType}).GroupBy("counterparty_account_id").ToSql()
	if err != nil {
		r.logger.ErrorContext(ctx, "building query", slog.Any("error", err))
		return nil, app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	rows, err := r.txc.DefaultTrOrDB(ctx, r.db).Query(ctx, query, args...)
	if err != nil {
		errIsConv, convErr := app.ErrConvertPgxToLogic(err)
		if !errIsConv {
			r.logger.ErrorContext(ctx, "executing query", slog.Any("error", err))
		}
		return nil, convErr
	}

	defer rows.Close()

	result := make([]usecase.CoinTransferRepositoryAggrHistoryItem, 0)

	for rows.Next() {
		data := CoinTransferAggrCoinHistoryItem{}
		if err := pgxscan.ScanRow(&data, rows); err != nil {
			errIsConv, convErr := app.ErrConvertPgxToLogic(err)
			if !errIsConv {
				r.logger.ErrorContext(ctx, "scan row", slog.Any("error", err))
			}
			return nil, convErr
		}
		result = append(result, usecase.CoinTransferRepositoryAggrHistoryItem{
			AccountID: data.AccountID,
			Ammount:   data.TotalAmount,
		})
	}
	if err := rows.Err(); err != nil {
		errIsConv, convErr := app.ErrConvertPgxToLogic(err)
		if !errIsConv {
			r.logger.ErrorContext(ctx, "scan row", slog.Any("error", err))
		}
		return nil, convErr
	}

	return result, nil
}
