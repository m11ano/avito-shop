package integration_test

import (
	"context"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/m11ano/avito-shop/internal/domain"
	"github.com/stretchr/testify/assert"
)

func (s *IntegrationTestSuite) TestCoinTransfer() {
	// Предварительно создадим 2 пользователей и дадим 100 монет
	initBalance := int64(100)

	testAccount1, err := domain.NewAccount("test1", "test")
	assert.NoError(s.T(), err)
	err = s.accountUsecase.Create(context.Background(), testAccount1)
	assert.NoError(s.T(), err)

	operation := domain.NewOperation(domain.OperationTypeIncrease, testAccount1.ID, initBalance, domain.OperationSourceTypeDeposit, nil)
	balance, err := s.operationUsecase.SaveOperation(context.Background(), operation)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), initBalance, balance)

	testAccount2, err := domain.NewAccount("test2", "test")
	assert.NoError(s.T(), err)
	err = s.accountUsecase.Create(context.Background(), testAccount2)
	assert.NoError(s.T(), err)

	operation = domain.NewOperation(domain.OperationTypeIncrease, testAccount2.ID, initBalance, domain.OperationSourceTypeDeposit, nil)
	balance, err = s.operationUsecase.SaveOperation(context.Background(), operation)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), initBalance, balance)

	// Предварительно авторизуемся под первым
	token, err := s.authUsecase.SignInOrSignUp(context.Background(), "test1", "test")
	assert.NoError(s.T(), err)

	testcases := []Case{
		// первый отправляет второму
		{
			name:       "success buy",
			timeoustMs: stdTimeout,
			request: Request{
				method: http.MethodPost,
				path:   "/api/sendCoin",
				headers: map[string]string{
					"Authorization": token,
				},
				body: map[string]interface{}{
					"toUser": "test2",
					"amount": 10,
				},
			},
			expectStatusCode: http.StatusOK,
			parseType:        CaseParseTypeText,
			respCheck: RespCheck{
				func(_ any) {
					var count int64
					var sum int64

					// Проверяем отправителя
					err = s.pgxpool.QueryRow(context.Background(), "SELECT COUNT(*) as count, COALESCE(SUM(amount), 0) as sum FROM coin_transfer WHERE owner_account_id = $1 AND counterparty_account_id = $2 AND transfer_type = 1", testAccount1.ID, testAccount2.ID).Scan(&count, &sum)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), int64(1), count)
					assert.Equal(s.T(), int64(10), sum)

					var balance int64
					err = s.pgxpool.QueryRow(context.Background(), "SELECT balance FROM operation_balance WHERE account_id = $1", testAccount1.ID).Scan(&balance)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), initBalance-10, balance)

					err = s.pgxpool.QueryRow(context.Background(), "SELECT COALESCE(SUM(amount), 0) as balance FROM operation WHERE account_id = $1", testAccount1.ID).Scan(&balance)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), initBalance-10, balance)

					// Проверяем получателя
					err = s.pgxpool.QueryRow(context.Background(), "SELECT COUNT(*) as count, COALESCE(SUM(amount), 0) as sum FROM coin_transfer WHERE owner_account_id = $2 AND counterparty_account_id = $1 AND transfer_type = 2", testAccount1.ID, testAccount2.ID).Scan(&count, &sum)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), int64(1), count)
					assert.Equal(s.T(), int64(10), sum)

					err = s.pgxpool.QueryRow(context.Background(), "SELECT balance FROM operation_balance WHERE account_id = $1", testAccount2.ID).Scan(&balance)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), initBalance+10, balance)

					err = s.pgxpool.QueryRow(context.Background(), "SELECT COALESCE(SUM(amount), 0) as balance FROM operation WHERE account_id = $1", testAccount2.ID).Scan(&balance)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), initBalance+10, balance)
				},
			},
		},
		// первый отправляет второму сумму больше своего баланса
		{
			name:       "success buy",
			timeoustMs: stdTimeout,
			request: Request{
				method: http.MethodPost,
				path:   "/api/sendCoin",
				headers: map[string]string{
					"Authorization": token,
				},
				body: map[string]interface{}{
					"toUser": "test2",
					"amount": 200,
				},
			},
			expectStatusCode: http.StatusBadRequest,
			parseType:        CaseParseTypeText,
			respCheck: RespCheck{
				func(_ any) {
					var count int64
					var sum int64

					// Проверяем отправителя
					err = s.pgxpool.QueryRow(context.Background(), "SELECT COUNT(*) as count, COALESCE(SUM(amount), 0) as sum FROM coin_transfer WHERE owner_account_id = $1 AND counterparty_account_id = $2 AND transfer_type = 1", testAccount1.ID, testAccount2.ID).Scan(&count, &sum)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), int64(1), count)
					assert.Equal(s.T(), int64(10), sum)

					var balance int64
					err = s.pgxpool.QueryRow(context.Background(), "SELECT balance FROM operation_balance WHERE account_id = $1", testAccount1.ID).Scan(&balance)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), initBalance-10, balance)

					err = s.pgxpool.QueryRow(context.Background(), "SELECT COALESCE(SUM(amount), 0) as balance FROM operation WHERE account_id = $1", testAccount1.ID).Scan(&balance)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), initBalance-10, balance)

					// Проверяем получателя
					err = s.pgxpool.QueryRow(context.Background(), "SELECT COUNT(*) as count, COALESCE(SUM(amount), 0) as sum FROM coin_transfer WHERE owner_account_id = $2 AND counterparty_account_id = $1 AND transfer_type = 2", testAccount1.ID, testAccount2.ID).Scan(&count, &sum)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), int64(1), count)
					assert.Equal(s.T(), int64(10), sum)

					err = s.pgxpool.QueryRow(context.Background(), "SELECT balance FROM operation_balance WHERE account_id = $1", testAccount2.ID).Scan(&balance)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), initBalance+10, balance)

					err = s.pgxpool.QueryRow(context.Background(), "SELECT COALESCE(SUM(amount), 0) as balance FROM operation WHERE account_id = $1", testAccount2.ID).Scan(&balance)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), initBalance+10, balance)
				},
			},
		},
		// некорректные данные на вход
		{
			name:       "send coin with incorrect input",
			timeoustMs: stdTimeout,
			request: Request{
				method: http.MethodPost,
				path:   "/api/sendCoin",
				headers: map[string]string{
					"Authorization": token,
				},
				body: map[string]interface{}{
					"toUser": "",
					"amount": 0,
				},
			},
			expectStatusCode: http.StatusBadRequest,
		},
		// отправка несуществующему пользователю
		{
			name:       "send coin to not exists user",
			timeoustMs: stdTimeout,
			request: Request{
				method: http.MethodPost,
				path:   "/api/sendCoin",
				headers: map[string]string{
					"Authorization": token,
				},
				body: map[string]interface{}{
					"toUser": "not_exists",
					"amount": 0,
				},
			},
			expectStatusCode: http.StatusBadRequest,
		},
		// без данных на вход
		{
			name:       "send coin without input",
			timeoustMs: stdTimeout,
			request: Request{
				method: http.MethodPost,
				path:   "/api/sendCoin",
				headers: map[string]string{
					"Authorization": token,
				},
			},
			expectStatusCode: http.StatusBadRequest,
		},
		// без токена авторизации
		{
			name:       "send coin without auth token",
			timeoustMs: stdTimeout,
			request: Request{
				method: http.MethodPost,
				path:   "/api/sendCoin",
			},
			expectStatusCode: http.StatusUnauthorized,
		},
	}

	for _, testcase := range testcases {
		s.execTestCase(testcase)
	}
}
