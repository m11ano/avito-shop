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
	"github.com/m11ano/avito-shop/internal/domain"
	"github.com/m11ano/avito-shop/pkg/dbhelper"
	"github.com/m11ano/avito-shop/pkg/e"
)

const (
	accountTable = "account"
)

type DBAccount struct {
	ID           uuid.UUID `db:"account_id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

var (
	accountTableFields = []string{}
	accountDBSchema    = &DBAccount{}
)

func init() {
	accountTableFields = dbhelper.ExtractDBFields(accountDBSchema)
}

type Account struct {
	logger *slog.Logger
	db     *pgxpool.Pool
	txc    *trmpgx.CtxGetter
	qb     squirrel.StatementBuilderType
}

func NewAccount(logger *slog.Logger, db *pgxpool.Pool, txc *trmpgx.CtxGetter) *Account {
	return &Account{
		logger: logger,
		db:     db,
		txc:    txc,
		qb:     squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *Account) dbToDomain(db *DBAccount) *domain.Account {
	return &domain.Account{
		ID:           db.ID,
		Username:     db.Username,
		PasswordHash: db.PasswordHash,
		CreatedAt:    db.CreatedAt,
		UpdatedAt:    db.UpdatedAt,
	}
}

func (r *Account) FindItemByUsername(ctx context.Context, username string) (*domain.Account, error) {
	query, args, err := r.qb.Select(accountTableFields...).From(accountTable).Where(squirrel.Eq{"username": username}).Limit(1).ToSql()
	if err != nil {
		r.logger.ErrorContext(ctx, "building query", slog.Any("error", err))
		return nil, e.NewErrorFrom(e.ErrInternal).Wrap(err)
	}

	rows, err := r.txc.DefaultTrOrDB(ctx, r.db).Query(ctx, query, args...)
	if err != nil {
		errIsConv, convErr := e.ErrConvertPgxToLogic(err)
		if !errIsConv {
			r.logger.ErrorContext(ctx, "executing query", slog.Any("error", err))
		}
		return nil, convErr
	}

	defer rows.Close()

	dbData := &DBAccount{}

	if err := pgxscan.ScanOne(dbData, rows); err != nil {
		errIsConv, convErr := e.ErrConvertPgxToLogic(err)
		if !errIsConv {
			r.logger.ErrorContext(ctx, "scan row", slog.Any("error", err))
		}
		return nil, convErr
	}

	item := r.dbToDomain(dbData)

	return item, nil
}

func (r *Account) FindItemsByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]domain.Account, error) {
	query, args, err := r.qb.Select(accountTableFields...).From(accountTable).Where(squirrel.Eq{"account_id": ids}).ToSql()
	if err != nil {
		r.logger.ErrorContext(ctx, "building query", slog.Any("error", err))
		return nil, e.NewErrorFrom(e.ErrInternal).Wrap(err)
	}

	rows, err := r.txc.DefaultTrOrDB(ctx, r.db).Query(ctx, query, args...)
	if err != nil {
		errIsConv, convErr := e.ErrConvertPgxToLogic(err)
		if !errIsConv {
			r.logger.ErrorContext(ctx, "executing query", slog.Any("error", err))
		}
		return nil, convErr
	}

	defer rows.Close()

	result := map[uuid.UUID]domain.Account{}

	for rows.Next() {
		data := &DBAccount{}
		if err := pgxscan.ScanRow(data, rows); err != nil {
			errIsConv, convErr := e.ErrConvertPgxToLogic(err)
			if !errIsConv {
				r.logger.ErrorContext(ctx, "scan row", slog.Any("error", err))
			}
			return nil, convErr
		}
		domainItem := *r.dbToDomain(data)
		result[domainItem.ID] = domainItem
	}
	if err := rows.Err(); err != nil {
		errIsConv, convErr := e.ErrConvertPgxToLogic(err)
		if !errIsConv {
			r.logger.ErrorContext(ctx, "scan row", slog.Any("error", err))
		}
		return nil, convErr
	}

	return result, nil
}

func (r *Account) Create(ctx context.Context, item *domain.Account) error {
	dataMap, err := dbhelper.StructToDBMap(item, accountDBSchema)
	if err != nil {
		r.logger.ErrorContext(ctx, "convert struct to db map", slog.Any("error", err))
		return e.NewErrorFrom(e.ErrInternal).Wrap(err)
	}
	dataMap["updated_at"] = time.Now()

	query, args, err := r.qb.Insert(accountTable).SetMap(dataMap).ToSql()
	if err != nil {
		r.logger.ErrorContext(ctx, "building query", slog.Any("error", err))
		return e.NewErrorFrom(e.ErrInternal).Wrap(err)
	}

	_, err = r.txc.DefaultTrOrDB(ctx, r.db).Exec(ctx, query, args...)
	if err != nil {
		errIsConv, convErr := e.ErrConvertPgxToLogic(err)
		if !errIsConv {
			r.logger.ErrorContext(ctx, "executing query", slog.Any("error", err))
		}
		return convErr
	}

	return nil
}
