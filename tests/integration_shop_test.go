package integration_test

import (
	"context"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/m11ano/avito-shop/internal/domain"
	"github.com/stretchr/testify/assert"
)

func (s *IntegrationTestSuite) TestShopBuy() {
	// Предварительно создадим пользователя и дадим 100 монет
	initBalance := int64(100)

	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)
	err = s.accountUsecase.Create(context.Background(), testAccount)
	assert.NoError(s.T(), err)

	operation := domain.NewOperation(domain.OperationTypeIncrease, testAccount.ID, initBalance, domain.OperationSourceTypeDeposit, nil)
	balance, err := s.operationUsecase.SaveOperation(context.Background(), operation)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), initBalance, balance)

	// Предварительно авторизуемся
	token, err := s.authUsecase.SignInOrSignUp(context.Background(), "test", "test")
	assert.NoError(s.T(), err)

	// Предварительно извлечем товары
	pen, err := s.shopItemUsecase.GetItemByName(context.Background(), "pen")
	assert.NoError(s.T(), err)

	hoody, err := s.shopItemUsecase.GetItemByName(context.Background(), "hoody")
	assert.NoError(s.T(), err)

	testcases := []Case{
		// успешная покупка
		{
			name:       "success buy",
			timeoustMs: stdTimeout,
			request: Request{
				method: http.MethodGet,
				path:   "/api/buy/pen",
				headers: map[string]string{
					"Authorization": token,
				},
			},
			expectStatusCode: http.StatusOK,
			parseType:        CaseParseTypeText,
			respCheck: RespCheck{
				func(_ any) {
					var count int64
					var quantity int64
					err = s.pgxpool.QueryRow(context.Background(), "SELECT COUNT(*) as count, COALESCE(SUM(quantity), 0) as q FROM shop_purchase WHERE account_id = $1 AND item_id = $2", testAccount.ID, pen.ID).Scan(&count, &quantity)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), int64(1), count)
					assert.Equal(s.T(), int64(1), quantity)

					var balance int64
					err = s.pgxpool.QueryRow(context.Background(), "SELECT balance FROM operation_balance WHERE account_id = $1", testAccount.ID).Scan(&balance)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), initBalance-pen.Price, balance)

					err = s.pgxpool.QueryRow(context.Background(), "SELECT COALESCE(SUM(amount), 0) as balance FROM operation WHERE account_id = $1", testAccount.ID).Scan(&balance)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), initBalance-pen.Price, balance)
				},
			},
		},
		// повторная успешная покупка
		{
			name:       "repeat success buy",
			timeoustMs: stdTimeout,
			request: Request{
				method: http.MethodGet,
				path:   "/api/buy/pen",
				headers: map[string]string{
					"Authorization": token,
				},
			},
			expectStatusCode: http.StatusOK,
			parseType:        CaseParseTypeText,
			respCheck: RespCheck{
				func(_ any) {
					var count int64
					var quantity int64
					err = s.pgxpool.QueryRow(context.Background(), "SELECT COUNT(*) as count, COALESCE(SUM(quantity), 0) as q FROM shop_purchase WHERE account_id = $1 AND item_id = $2", testAccount.ID, pen.ID).Scan(&count, &quantity)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), int64(2), count)
					assert.Equal(s.T(), int64(2), quantity)

					var balance int64
					err = s.pgxpool.QueryRow(context.Background(), "SELECT balance FROM operation_balance WHERE account_id = $1", testAccount.ID).Scan(&balance)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), initBalance-pen.Price*2, balance)

					err = s.pgxpool.QueryRow(context.Background(), "SELECT COALESCE(SUM(amount), 0) as balance FROM operation WHERE account_id = $1", testAccount.ID).Scan(&balance)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), initBalance-pen.Price*2, balance)
				},
			},
		},
		// недостаточно денег
		{
			name:       "buy shop item error: not enough money",
			timeoustMs: stdTimeout,
			request: Request{
				method: http.MethodGet,
				path:   "/api/buy/hoody",
				headers: map[string]string{
					"Authorization": token,
				},
			},
			expectStatusCode: http.StatusBadRequest,
			parseType:        CaseParseTypeText,
			respCheck: RespCheck{
				func(_ any) {
					var count int64
					var quantity int64
					err = s.pgxpool.QueryRow(context.Background(), "SELECT COUNT(*) as count, COALESCE(SUM(quantity), 0) as q FROM shop_purchase WHERE account_id = $1 AND item_id = $2", testAccount.ID, hoody.ID).Scan(&count, &quantity)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), int64(0), count)
					assert.Equal(s.T(), int64(0), quantity)

					// Проверим что баланс не изменился с прошлого раза
					var balance int64
					err = s.pgxpool.QueryRow(context.Background(), "SELECT balance FROM operation_balance WHERE account_id = $1", testAccount.ID).Scan(&balance)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), initBalance-pen.Price*2, balance)

					err = s.pgxpool.QueryRow(context.Background(), "SELECT COALESCE(SUM(amount), 0) as balance FROM operation WHERE account_id = $1", testAccount.ID).Scan(&balance)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), initBalance-pen.Price*2, balance)
				},
			},
		},
		// без токена авторизации
		{
			name:       "buy shop item without auth token",
			timeoustMs: stdTimeout,
			request: Request{
				method: http.MethodGet,
				path:   "/api/buy/pen",
			},
			expectStatusCode: http.StatusUnauthorized,
		},
		// некорректный товар
		{
			name:       "buy inccorect shop item",
			timeoustMs: stdTimeout,
			request: Request{
				method: http.MethodGet,
				path:   "/api/buy/notexists",
				headers: map[string]string{
					"Authorization": token,
				},
			},
			expectStatusCode: http.StatusBadRequest,
		},
	}

	for _, testcase := range testcases {
		s.execTestCase(testcase)
	}
}
