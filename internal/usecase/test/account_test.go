package usecase_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/internal/app"
	"github.com/m11ano/avito-shop/internal/config"
	"github.com/m11ano/avito-shop/internal/db/txmngr"
	"github.com/m11ano/avito-shop/internal/domain"
	"github.com/m11ano/avito-shop/internal/usecase"
	"github.com/m11ano/avito-shop/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

type AccountTestSuite struct {
	suite.Suite
	mockPgxPool    *mocks.PgxPool
	accountService usecase.Account
	app            *fx.App
	mockRepo       *mocks.AccountRepository
}

func (s *AccountTestSuite) SetupTest() {
	s.mockPgxPool = mocks.NewPgxPoolMockForTxManager()
	s.mockRepo = new(mocks.AccountRepository)

	s.app = fx.New(
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
		fx.Provide(func() config.Config { return config.LoadConfig("../../../config.yml") }),
		fx.Provide(func() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }),
		fx.Provide(txmngr.NewProvider(s.mockPgxPool)),
		fx.Provide(func() usecase.AccountRepository { return s.mockRepo }),
		fx.Provide(fx.Annotate(usecase.NewAccountInpl, fx.As(new(usecase.Account)))),
		fx.Populate(&s.accountService),
	)

	err := s.app.Start(context.Background())
	assert.NoError(s.T(), err)
}

func (s *AccountTestSuite) TearDownTest() {
	err := s.app.Stop(context.Background())
	assert.NoError(s.T(), err)
}

func TestAccountSuiteRun(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}

func (s *AccountTestSuite) TestGetItemByUsername__OK() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	s.mockRepo.On("FindItemByUsername", mock.Anything, "test").Return(testAccount, nil)

	account, err := s.accountService.GetItemByUsername(context.Background(), "test")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), testAccount, account)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *AccountTestSuite) TestGetItemByUsername__NotFound() {
	s.mockRepo.On("FindItemByUsername", mock.Anything, "test").Return(nil, app.ErrNotFound)

	account, err := s.accountService.GetItemByUsername(context.Background(), "test")
	assert.ErrorIs(s.T(), err, app.ErrNotFound)
	assert.Nil(s.T(), account)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *AccountTestSuite) TestGetItemByIDs__OK() {
	testAccount1, err := domain.NewAccount("test_1", "test")
	assert.NoError(s.T(), err)
	testAccount2, err := domain.NewAccount("test_2", "test")
	assert.NoError(s.T(), err)

	s.mockRepo.On("FindItemsByIDs", mock.Anything, []uuid.UUID{testAccount1.ID, testAccount2.ID}).Return(map[uuid.UUID]domain.Account{testAccount1.ID: *testAccount1, testAccount2.ID: *testAccount2}, nil)

	accounts, err := s.accountService.GetItemsByIDs(context.Background(), []uuid.UUID{testAccount1.ID, testAccount2.ID})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), map[uuid.UUID]domain.Account{testAccount1.ID: *testAccount1, testAccount2.ID: *testAccount2}, accounts)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *AccountTestSuite) TestCreate__OK() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	s.mockRepo.On("Create", mock.Anything, testAccount).Return(nil)

	testAccountCopy := *testAccount // not deep copy
	err = s.accountService.Create(context.Background(), testAccount)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), &testAccountCopy, testAccount)

	s.mockRepo.AssertExpectations(s.T())
}
