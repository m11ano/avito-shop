package app

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func ErrCheckIsTxСoncurrentExec(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "40001" {
		return true
	}
	return errors.Is(err, ErrTxСoncurrentExec)
}

func ErrConvertPgxToLogic(err error) (bool, error) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch {
		case pgErr.Code == "40001":
			return true, NewErrorFrom(ErrTxСoncurrentExec).Wrap(err)
		case pgErr.Code == "23505":
			return true, NewErrorFrom(ErrStoreUniqueViolation).Wrap(err).SetData(pgErr.ColumnName)
		case pgErr.Code == "23503":
			return true, NewErrorFrom(ErrStoreForeignKeyViolation).Wrap(err).SetData(pgErr.ColumnName)
		case pgErr.Code == "23502":
			return true, NewErrorFrom(ErrStoreNotNullViolation).Wrap(err).SetData(pgErr.ColumnName)
		case pgErr.Code == "23514":
			return true, NewErrorFrom(ErrStoreCheckViolation).Wrap(err).SetData(pgErr.ConstraintName)
		case pgErr.Code == "23001":
			return true, NewErrorFrom(ErrStoreRestrictViolation).Wrap(err).SetData(pgErr.ConstraintName)
		case pgErr.Code == "23000":
			return true, NewErrorFrom(ErrStoreIntegrityViolation).Wrap(err).SetData(pgErr.ConstraintName)
		default:
			return false, NewErrorFrom(ErrInternal).Wrap(err)
		}
	}
	return false, err
}
