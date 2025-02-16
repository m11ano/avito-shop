package integration_test

import (
	"context"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/m11ano/avito-shop/internal/domain"
	"github.com/stretchr/testify/assert"
)

type InfoHandlerOut struct {
	Coins       int64                     `json:"coins"`
	Inventory   []InfoHandlerOutInventory `json:"inventory"`
	CoinHistory InfoHandlerOutCoinHistory `json:"coinHistory"`
}

type InfoHandlerOutInventory struct {
	Type     string `json:"type"`
	Quantity int64  `json:"quantity"`
}

type InfoHandlerOutCoinHistory struct {
	Received []InfoHandlerOutCoinHistoryReceived `json:"received"`
	Sent     []InfoHandlerOutCoinHistorySent     `json:"sent"`
}

type InfoHandlerOutCoinHistorySent struct {
	ToUser string `json:"toUser"`
	Amount int64  `json:"amount"`
}

type InfoHandlerOutCoinHistoryReceived struct {
	FromUser string `json:"fromUser"`
	Amount   int64  `json:"amount"`
}

func (s *IntegrationTestSuite) TestInfo() {
	// Предварительно создадим 2 пользователей и дадим 1000 монет
	initBalance := int64(1000)

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

	// Предварительно извлечем товары
	pen, err := s.shopItemUsecase.GetItemByName(context.Background(), "pen")
	assert.NoError(s.T(), err)

	// Предварительно сделаем покупки в магазине
	_, err = s.shopPurchaseUsecase.MakePurchase(context.Background(), pen.Name, testAccount1.ID, 1, nil)
	assert.NoError(s.T(), err)

	// Предварительно делаем переводы между пользователя
	_, _, err = s.coinTransferUsecase.MakeTransferByUsername(context.Background(), testAccount2.Username, testAccount1.ID, 10, nil)
	assert.NoError(s.T(), err)
	_, _, err = s.coinTransferUsecase.MakeTransferByUsername(context.Background(), testAccount1.Username, testAccount2.ID, 50, nil)
	assert.NoError(s.T(), err)

	// Предварительно авторизуемся под первым
	token, err := s.authUsecase.SignInOrSignUp(context.Background(), "test1", "test")
	assert.NoError(s.T(), err)

	testcases := []Case{
		// основной запрос на инфо
		{
			name:       "success info",
			timeoustMs: stdTimeout,
			request: Request{
				method: http.MethodGet,
				path:   "/api/info",
				headers: map[string]string{
					"Authorization": token,
				},
			},
			expectStatusCode: http.StatusOK,
			parseType:        CaseParseTypeJSON,
			parseResponse:    &InfoHandlerOut{},
			respCheck: RespCheck{
				func(resp any) {
					respParsed, ok := resp.(*InfoHandlerOut)
					assert.True(s.T(), ok, "Response should be of type *InfoHandlerOut")

					assert.Equal(s.T(), initBalance-pen.Price-10+50, respParsed.Coins)

					assert.Equal(s.T(), []InfoHandlerOutInventory{
						{
							Type:     pen.Name,
							Quantity: 1,
						},
					}, respParsed.Inventory)

					assert.Equal(s.T(), []InfoHandlerOutCoinHistoryReceived{
						{
							FromUser: "test2",
							Amount:   50,
						},
					}, respParsed.CoinHistory.Received)

					assert.Equal(s.T(), []InfoHandlerOutCoinHistorySent{
						{
							ToUser: "test2",
							Amount: 10,
						},
					}, respParsed.CoinHistory.Sent)
				},
			},
		},
		// без токена авторизации
		{
			name:       "info without auth token",
			timeoustMs: stdTimeout,
			request: Request{
				method: http.MethodGet,
				path:   "/api/info",
			},
			expectStatusCode: http.StatusUnauthorized,
		},
	}

	for _, testcase := range testcases {
		s.execTestCase(testcase)
	}
}
