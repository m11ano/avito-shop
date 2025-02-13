package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Masterminds/squirrel"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m11ano/avito-shop/internal/app"
	"github.com/m11ano/avito-shop/internal/domain"
	"github.com/m11ano/avito-shop/pkg/dbhelper"
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
	accountTableFields = dbhelper.ExtractDBFields(DBAccount{})
}

type Account struct {
	logger    *slog.Logger
	db        *pgxpool.Pool
	txc       *trmpgx.CtxGetter
	qb        squirrel.StatementBuilderType
	txManager *manager.Manager
}

func NewAccount(logger *slog.Logger, db *pgxpool.Pool, txc *trmpgx.CtxGetter, txManager *manager.Manager) *Account {
	return &Account{
		logger:    logger,
		db:        db,
		txc:       txc,
		qb:        squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		txManager: txManager,
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
		return nil, app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	rows, err := r.txc.DefaultTrOrDB(ctx, r.db).Query(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "40001" {
			return nil, app.NewErrorFrom(app.ErrTxСoncurrentExec).Wrap(err)
		}
		r.logger.ErrorContext(ctx, "executing query", "error", err)
		return nil, app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	defer rows.Close()

	dbData := &DBAccount{}

	if err := pgxscan.ScanOne(dbData, rows); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app.NewErrorFrom(app.ErrNotFound).Wrap(err)
		}
		r.logger.ErrorContext(ctx, "scan row", slog.Any("error", err))
		return nil, app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	item := r.dbToDomain(dbData)

	return item, nil
}

func (r *Account) Create(ctx context.Context, item *domain.Account) error {
	dataMap, err := dbhelper.StructToDBMap(item, accountDBSchema)
	if err != nil {
		r.logger.ErrorContext(ctx, "convert struct to db map", slog.Any("error", err))
		return app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}
	dataMap["created_at"] = time.Now()
	dataMap["updated_at"] = time.Now()

	query, args, err := r.qb.Insert(accountTable).SetMap(dataMap).ToSql()
	if err != nil {
		r.logger.ErrorContext(ctx, "building query", slog.Any("error", err))
		return app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	_, err = r.txc.DefaultTrOrDB(ctx, r.db).Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "40001" {
				return app.NewErrorFrom(app.ErrTxСoncurrentExec).Wrap(err)
			}
			if pgErr.Code == "23505" {
				return app.NewErrorFrom(app.ErrUniqueViolation).Wrap(err).SetData(pgErr.ColumnName)
			}
		}
		r.logger.ErrorContext(ctx, "executing query", slog.Any("error", err))
		return app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	return nil
}

func (r *Account) Update(ctx context.Context, item *domain.Account, id uuid.UUID) error {
	dataMap, err := dbhelper.StructToDBMap(item, accountDBSchema)
	if err != nil {
		r.logger.ErrorContext(ctx, "convert struct to db map", slog.Any("error", err))
		return app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}
	delete(dataMap, "created_at")
	dataMap["updated_at"] = time.Now()

	query, args, err := r.qb.Update(accountTable).SetMap(dataMap).Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		r.logger.ErrorContext(ctx, "error building query", slog.Any("error", err))
		return app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	_, err = r.txc.DefaultTrOrDB(ctx, r.db).Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "40001" {
			return app.NewErrorFrom(app.ErrTxСoncurrentExec).Wrap(err)
		}
		r.logger.ErrorContext(ctx, "error executing query", slog.Any("error", err))
		return app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	return nil
}
