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
	"github.com/m11ano/avito-shop/pkg/dbhelper"
)

const (
	operationTable = "operation"
)

type DBOperation struct {
	ID         uuid.UUID                  `db:"operation_id"`
	Type       domain.OperationType       `db:"operation_type"`
	AccountID  uuid.UUID                  `db:"account_id"`
	Amount     int64                      `db:"amount"`
	SourceType domain.OperationSourceType `db:"source_type"`
	SourceID   *uuid.UUID                 `db:"source_id"`
	CreatedAt  time.Time                  `db:"created_at"`
}

var (
	//nolint:unused
	operationTableFields = []string{}
	operationDBSchema    = &DBOperation{}
)

func init() {
	operationTableFields = dbhelper.ExtractDBFields(operationDBSchema)
}

type Operation struct {
	logger *slog.Logger
	db     *pgxpool.Pool
	txc    *trmpgx.CtxGetter
	qb     squirrel.StatementBuilderType
}

func NewOperation(logger *slog.Logger, db *pgxpool.Pool, txc *trmpgx.CtxGetter) *Operation {
	return &Operation{
		logger: logger,
		db:     db,
		txc:    txc,
		qb:     squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

//nolint:unused
func (r *Operation) dbToDomain(db *DBOperation) *domain.Operation {
	item := &domain.Operation{
		ID:         db.ID,
		Type:       db.Type,
		AccountID:  db.AccountID,
		Amount:     db.Amount,
		SourceType: db.SourceType,
		SourceID:   db.SourceID,
		CreatedAt:  db.CreatedAt,
	}
	if item.Type == domain.OperationTypeDecrease {
		item.Amount = -item.Amount
	}
	return item
}

func (r *Operation) domainToDB(item *domain.Operation) (map[string]interface{}, error) {
	dataMap, err := dbhelper.StructToDBMap(item, operationDBSchema)
	if err != nil {
		return nil, app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	if item.Type == domain.OperationTypeDecrease {
		dataMap["amount"] = -item.Amount
	}

	return dataMap, nil
}

type CountBalanceDTO struct {
	Balance        int64 `db:"balance"`
	OperationCount int64 `db:"operation_count"`
}

func (r *Operation) CountBalanceByAccountID(ctx context.Context, accountID uuid.UUID) (int64, int64, error) {
	query, args, err := r.qb.Select("COALESCE(SUM(amount), 0) AS balance", "COUNT(*) as operation_count").From(operationTable).Where(squirrel.Eq{"account_id": accountID}).ToSql()
	if err != nil {
		r.logger.ErrorContext(ctx, "building query", slog.Any("error", err))
		return 0, 0, app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	rows, err := r.txc.DefaultTrOrDB(ctx, r.db).Query(ctx, query, args...)
	if err != nil {
		errIsConv, convErr := app.ErrConvertPgxToLogic(err)
		if !errIsConv {
			r.logger.ErrorContext(ctx, "executing query", slog.Any("error", err))
		}
		return 0, 0, convErr
	}

	defer rows.Close()

	dbData := &CountBalanceDTO{}

	if err := pgxscan.ScanOne(dbData, rows); err != nil {
		r.logger.ErrorContext(ctx, "scan row", slog.Any("error", err))
		return 0, 0, app.NewErrorFrom(app.ErrInternal).Wrap(err)
	}

	return dbData.Balance, dbData.OperationCount, nil
}

func (r *Operation) Create(ctx context.Context, item *domain.Operation) error {
	dataMap, err := r.domainToDB(item)
	if err != nil {
		r.logger.ErrorContext(ctx, "convert struct to db map", slog.Any("error", err))
		return err
	}

	query, args, err := r.qb.Insert(operationTable).SetMap(dataMap).ToSql()
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
