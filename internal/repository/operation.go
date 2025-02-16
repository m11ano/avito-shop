package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Masterminds/squirrel"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m11ano/avito-shop/internal/domain"
	"github.com/m11ano/avito-shop/pkg/dbhelper"
	"github.com/m11ano/avito-shop/pkg/e"
)

const (
	operationTable        = "operation"
	operationBalanceTable = "operation_balance"
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
	logger    *slog.Logger
	db        *pgxpool.Pool
	txc       *trmpgx.CtxGetter
	qb        squirrel.StatementBuilderType
	txManager *manager.Manager
}

func NewOperation(logger *slog.Logger, db *pgxpool.Pool, txc *trmpgx.CtxGetter, txManager *manager.Manager) *Operation {
	return &Operation{
		logger:    logger,
		db:        db,
		txc:       txc,
		qb:        squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		txManager: txManager,
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
		return nil, e.NewErrorFrom(e.ErrInternal).Wrap(err)
	}

	if item.Type == domain.OperationTypeDecrease {
		dataMap["amount"] = -item.Amount
	}

	return dataMap, nil
}

type OperationGetBalanceDTO struct {
	Balance int64 `db:"balance"`
}

func (r *Operation) GetBalanceByAccountID(ctx context.Context, accountID uuid.UUID) (int64, bool, error) {
	query, args, err := r.qb.Select("balance").From(operationBalanceTable).Where(squirrel.Eq{"account_id": accountID}).Limit(1).ToSql()
	if err != nil {
		r.logger.ErrorContext(ctx, "building query", slog.Any("error", err))
		return 0, false, e.NewErrorFrom(e.ErrInternal).Wrap(err)
	}

	rows, err := r.txc.DefaultTrOrDB(ctx, r.db).Query(ctx, query, args...)
	if err != nil {
		errIsConv, convErr := e.ErrConvertPgxToLogic(err)
		if !errIsConv {
			r.logger.ErrorContext(ctx, "executing query", slog.Any("error", err))
		}
		return 0, false, convErr
	}

	defer rows.Close()

	dbData := &OperationGetBalanceDTO{}

	if err := pgxscan.ScanOne(dbData, rows); err != nil {
		errIsConv, convErr := e.ErrConvertPgxToLogic(err)
		if !errIsConv {
			r.logger.ErrorContext(ctx, "scan row", slog.Any("error", err))
		}
		if errors.Is(convErr, e.ErrStoreNoRows) {
			return 0, false, nil
		}
		return 0, false, convErr
	}

	return dbData.Balance, true, nil
}

// Внутренний метод, обновить баланс по логу операций
func (r *Operation) updateBalanceByAccountID(ctx context.Context, accountID uuid.UUID) error {
	updateQuery := squirrel.Expr("INSERT INTO operation_balance (account_id, balance) VALUES ($1, (SELECT COALESCE(SUM(amount), 0) FROM operation WHERE account_id = $1)) ON CONFLICT (account_id) DO UPDATE SET balance = (SELECT COALESCE(SUM(amount), 0) FROM operation WHERE account_id = $1)", accountID)

	query, args, err := updateQuery.ToSql()
	if err != nil {
		r.logger.ErrorContext(ctx, "building query", slog.Any("error", err))
		return e.NewErrorFrom(e.ErrInternal).Wrap(err)
	}

	rows, err := r.txc.DefaultTrOrDB(ctx, r.db).Query(ctx, query, args...)
	if err != nil {
		fmt.Println("CATCH")
		errIsConv, convErr := e.ErrConvertPgxToLogic(err)
		if !errIsConv {
			r.logger.ErrorContext(ctx, "executing query", slog.Any("error", err))
		}
		return convErr
	}

	defer rows.Close()

	return nil
}

func (r *Operation) Create(ctx context.Context, item *domain.Operation) (int64, error) {
	dataMap, err := r.domainToDB(item)
	if err != nil {
		r.logger.ErrorContext(ctx, "convert struct to db map", slog.Any("error", err))
		return 0, err
	}

	var balance int64

	// DOTO: подумать целесообразности ретраев при возникновении блокировки
	err = r.txManager.Do(ctx, func(ctx context.Context) error {
		query, args, err := r.qb.Insert(operationTable).SetMap(dataMap).ToSql()
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

		// Если две конкурирующие транзакции попытаются обновить баланс аккаунта - одна из них получит блокировку
		err = r.updateBalanceByAccountID(ctx, item.AccountID)
		if err != nil {
			return err
		}

		// Далее если транзакция получила блокировку, здесь при попытке чтения вызовется ошибка
		balance, _, err = r.GetBalanceByAccountID(ctx, item.AccountID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return balance, nil
}
