package mocks

import (
	mock "github.com/stretchr/testify/mock"
)

func NewPgxPoolMockForTxManager() *PgxPool {
	mockPool := new(PgxPool)
	mockTx := new(PoolTxInterface)

	// Настраиваем пул: при вызове BeginTx(...) он вернёт mockTx
	mockPool.On("BeginTx", mock.Anything, mock.AnythingOfType("pgx.TxOptions")).Return(mockTx, nil)

	// Настраиваем транзакцию: Commit и Rollback возвращают nil (успех)
	mockTx.On("Commit", mock.Anything).Return(nil)
	mockTx.On("Rollback", mock.Anything).Return(nil)

	return mockPool
}
