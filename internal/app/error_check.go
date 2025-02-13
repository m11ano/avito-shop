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
