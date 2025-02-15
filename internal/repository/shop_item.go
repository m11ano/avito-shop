package repository

import (
	"context"
	"log/slog"

	"github.com/Masterminds/squirrel"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m11ano/avito-shop/internal/app"
	"github.com/m11ano/avito-shop/internal/domain"
	"github.com/m11ano/avito-shop/pkg/dbhelper"
)

const (
	shopItemTable = "shop_item"
)

type DBShopItem struct {
	ID    uuid.UUID `db:"item_id"`
	Name  string    `db:"item_name"`
	Price int64     `db:"price"`
}

var (
	shopItemTableFields = []string{}
	shopItemDBSchema    = &DBShopItem{}
)

func init() {
	shopItemTableFields = dbhelper.ExtractDBFields(shopItemDBSchema)
}

type ShopItem struct {
	logger *slog.Logger
	db     *pgxpool.Pool
	txc    *trmpgx.CtxGetter
	qb     squirrel.StatementBuilderType
}

func NewShopItem(logger *slog.Logger, db *pgxpool.Pool, txc *trmpgx.CtxGetter) *ShopItem {
	return &ShopItem{
		logger: logger,
		db:     db,
		txc:    txc,
		qb:     squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *ShopItem) dbToDomain(db *DBShopItem) *domain.ShopItem {
	return &domain.ShopItem{
		ID:    db.ID,
		Name:  db.Name,
		Price: db.Price,
	}
}

func (r *ShopItem) FindItemByID(ctx context.Context, id uuid.UUID) (*domain.ShopItem, error) {
	query, args, err := r.qb.Select(shopItemTableFields...).From(shopItemTable).Where(squirrel.Eq{"item_id": id}).Limit(1).ToSql()
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

	dbData := &DBShopItem{}

	if err := pgxscan.ScanOne(dbData, rows); err != nil {
		errIsConv, convErr := app.ErrConvertPgxToLogic(err)
		if !errIsConv {
			r.logger.ErrorContext(ctx, "scan row", slog.Any("error", err))
		}
		return nil, convErr
	}

	item := r.dbToDomain(dbData)

	return item, nil
}

func (r *ShopItem) FindItemByName(ctx context.Context, name string) (*domain.ShopItem, error) {
	query, args, err := r.qb.Select(shopItemTableFields...).From(shopItemTable).Where(squirrel.Eq{"item_name": name}).Limit(1).ToSql()
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

	dbData := &DBShopItem{}

	if err := pgxscan.ScanOne(dbData, rows); err != nil {
		errIsConv, convErr := app.ErrConvertPgxToLogic(err)
		if !errIsConv {
			r.logger.ErrorContext(ctx, "scan row", slog.Any("error", err))
		}
		return nil, convErr
	}

	item := r.dbToDomain(dbData)

	return item, nil
}

func (r *ShopItem) FindItemsByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.ShopItem, error) {
	query, args, err := r.qb.Select(shopItemTableFields...).From(shopItemTable).Where(squirrel.Eq{"item_id": ids}).ToSql()
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

	result := make([]domain.ShopItem, 0)

	for rows.Next() {
		data := &DBShopItem{}
		if err := pgxscan.ScanRow(data, rows); err != nil {
			errIsConv, convErr := app.ErrConvertPgxToLogic(err)
			if !errIsConv {
				r.logger.ErrorContext(ctx, "scan row", slog.Any("error", err))
			}
			return nil, convErr
		}
		result = append(result, *r.dbToDomain(data))
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
