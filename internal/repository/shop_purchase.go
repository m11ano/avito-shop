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
	shopPurchaseTable = "shop_purchase"
)

type DBShopPurchase struct {
	ID          uuid.UUID  `db:"purchase_id"`
	ItemID      uuid.UUID  `db:"item_id"`
	AccountID   uuid.UUID  `db:"account_id"`
	Quantity    int64      `db:"quantity"`
	CreatedAt   time.Time  `db:"created_at"`
	IdentityKey *uuid.UUID `db:"identity_key"`
}

var (
	shopPurchaseTableFields = []string{}
	shopPurchaseDBSchema    = &DBShopPurchase{}
)

func init() {
	shopPurchaseTableFields = dbhelper.ExtractDBFields(shopPurchaseDBSchema)
}

type ShopPurchase struct {
	logger *slog.Logger
	db     *pgxpool.Pool
	txc    *trmpgx.CtxGetter
	qb     squirrel.StatementBuilderType
}

func NewShopPurchase(logger *slog.Logger, db *pgxpool.Pool, txc *trmpgx.CtxGetter) *ShopPurchase {
	return &ShopPurchase{
		logger: logger,
		db:     db,
		txc:    txc,
		qb:     squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *ShopPurchase) dbToDomain(db *DBShopPurchase) *domain.ShopPurchase {
	return &domain.ShopPurchase{
		ID:          db.ID,
		ItemID:      db.ItemID,
		AccountID:   db.AccountID,
		Quantity:    db.Quantity,
		CreatedAt:   db.CreatedAt,
		IdentityKey: db.IdentityKey,
	}
}

type CheckIdentityDTO struct {
	Count int `db:"count"`
}

func (r *ShopPurchase) FindIdentity(ctx context.Context, key uuid.UUID) (bool, error) {
	query, args, err := r.qb.Select("COUNT(*) as count").From(shopPurchaseTable).Where(squirrel.Eq{"identity_key": key}).Limit(1).ToSql()
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

	dbData := &CheckIdentityDTO{}

	if err := pgxscan.ScanOne(dbData, rows); err != nil {
		r.logger.ErrorContext(ctx, "scan row", slog.Any("error", err))
		return false, app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	return dbData.Count > 0, nil
}

func (r *ShopPurchase) Create(ctx context.Context, item *domain.ShopPurchase) error {
	dataMap, err := dbhelper.StructToDBMap(item, shopPurchaseDBSchema)
	if err != nil {
		r.logger.ErrorContext(ctx, "convert struct to db map", slog.Any("error", err))
		return app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	query, args, err := r.qb.Insert(shopPurchaseTable).SetMap(dataMap).ToSql()
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

type AggrInventoryByAccountID struct {
	ItemID        uuid.UUID `db:"item_id"`
	TotalQuantity int64     `db:"total_quantity"`
}

func (r *ShopPurchase) AggrInventoryByAccountID(ctx context.Context, accountID uuid.UUID) ([]usecase.ShopPurchaseRepositoryAggrInventoryItem, error) {
	query, args, err := r.qb.Select("item_id", "COUNT(quantity) as total_quantity").From(shopPurchaseTable).Where(squirrel.Eq{"account_id": accountID}).GroupBy("item_id").ToSql()
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

	result := make([]usecase.ShopPurchaseRepositoryAggrInventoryItem, 0)

	for rows.Next() {
		data := AggrInventoryByAccountID{}
		if err := pgxscan.ScanRow(&data, rows); err != nil {
			r.logger.ErrorContext(ctx, "scan row", slog.Any("error", err))
			return nil, app.NewErrorFrom(app.ErrInternal).Wrap(err)
		}
		result = append(result, usecase.ShopPurchaseRepositoryAggrInventoryItem{
			ShopItemID: data.ItemID,
			Quantity:   data.TotalQuantity,
		})
	}
	if err := rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "scan row", slog.Any("error", err))
		return nil, app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	return result, nil
}
