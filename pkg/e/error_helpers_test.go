package e

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestErrCheckIsTxСoncurrentExec_PgError40001(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "40001"}
	assert.True(t, ErrCheckIsTxСoncurrentExec(pgErr))
}

func TestErrCheckIsTxСoncurrentExec_PgError25P02(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "25P02"}
	assert.True(t, ErrCheckIsTxСoncurrentExec(pgErr))
}

func TestErrCheckIsTxСoncurrentExec_IsErrTxConcurrent(t *testing.T) {
	err := fmt.Errorf("dummy wrap: %w", ErrTxСoncurrentExec)
	assert.True(t, ErrCheckIsTxСoncurrentExec(err))
}

func TestErrCheckIsTxСoncurrentExec_False(t *testing.T) {
	stdErr := errors.New("some other error")
	assert.False(t, ErrCheckIsTxСoncurrentExec(stdErr))
}

func TestErrConvertPgxToLogic_NoRows(t *testing.T) {
	isLogic, logicErr := ErrConvertPgxToLogic(pgx.ErrNoRows)
	assert.True(t, isLogic)
	assert.True(t, errors.Is(logicErr, ErrStoreNoRows))
	assert.True(t, errors.Is(logicErr, pgx.ErrNoRows))
}

func TestErrConvertPgxToLogic_PgErrors(t *testing.T) {
	tests := []struct {
		code               string
		wantWrapErr        *LogicError
		wantDataFromColumn bool
		wantDataFromConstr bool
	}{
		{"40001", ErrTxСoncurrentExec, false, false},
		{"25P02", ErrTxСoncurrentExec, false, false},
		{"23505", ErrStoreUniqueViolation, true, false},
		{"23503", ErrStoreForeignKeyViolation, true, false},
		{"23502", ErrStoreNotNullViolation, true, false},
		{"23514", ErrStoreCheckViolation, false, true},
		{"23001", ErrStoreRestrictViolation, false, true},
		{"23000", ErrStoreIntegrityViolation, false, true},
	}

	for _, tt := range tests {
		t.Run("code="+tt.code, func(t *testing.T) {
			pgErr := &pgconn.PgError{
				Code:           tt.code,
				ColumnName:     "test_column",
				ConstraintName: "test_constraint",
			}
			isLogic, logicErr := ErrConvertPgxToLogic(pgErr)
			assert.True(t, isLogic, "For code %s should return logic error", tt.code)
			assert.True(t, errors.Is(logicErr, tt.wantWrapErr), "Should wrap the specific logic error for %s", tt.code)
			assert.True(t, errors.Is(logicErr, pgErr), "Should wrap original pgErr for %s", tt.code)

			appErr, ok := logicErr.(*LogicError)
			assert.True(t, ok)
			if tt.wantDataFromColumn {
				assert.Equal(t, "test_column", appErr.Data())
			} else if tt.wantDataFromConstr {
				assert.Equal(t, "test_constraint", appErr.Data())
			} else {
				assert.Nil(t, appErr.Data())
			}
		})
	}
}

func TestErrConvertPgxToLogic_PgErrorDefaultCase(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "99999"}
	isLogic, logicErr := ErrConvertPgxToLogic(pgErr)

	assert.False(t, isLogic)
	assert.True(t, errors.Is(logicErr, ErrInternal))
	assert.True(t, errors.Is(logicErr, pgErr))
}

func TestErrConvertPgxToLogic_NoPgError(t *testing.T) {
	stdErr := errors.New("some random error")
	isLogic, err := ErrConvertPgxToLogic(stdErr)
	assert.False(t, isLogic)
	assert.Equal(t, stdErr, err)
}
