package usecase_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/m11ano/avito-shop/internal/config"
	"github.com/m11ano/avito-shop/internal/db/txmngr"
	"github.com/m11ano/avito-shop/internal/domain"
	"github.com/m11ano/avito-shop/internal/usecase"
	"github.com/m11ano/avito-shop/pkg/e"
	"github.com/m11ano/avito-shop/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

type OperationTestSuite struct {
	suite.Suite
	mockPgxPool      *mocks.PgxPool
	operationUsecase usecase.Operation
	app              *fx.App
	mockRepo         *mocks.OperationRepository
}

func (s *OperationTestSuite) SetupTest() {
	s.mockPgxPool = mocks.NewPgxPoolMockForTxManager()
	s.mockRepo = new(mocks.OperationRepository)

	s.app = fx.New(
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
		fx.Provide(func() config.Config { return config.LoadConfig("../../../config.yml") }),
		fx.Provide(func() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }),
		fx.Provide(txmngr.NewProvider(s.mockPgxPool)),
		fx.Provide(func() usecase.OperationRepository { return s.mockRepo }),
		fx.Provide(fx.Annotate(usecase.NewOperationInpl, fx.As(new(usecase.Operation)))),
		fx.Populate(&s.operationUsecase),
	)

	err := s.app.Start(context.Background())
	assert.NoError(s.T(), err)
}

func (s *OperationTestSuite) TearDownTest() {
	err := s.app.Stop(context.Background())
	assert.NoError(s.T(), err)
}

func TestOperationSuiteRun(t *testing.T) {
	suite.Run(t, new(OperationTestSuite))
}

func (s *OperationTestSuite) TestGetBalanceByAccountID__OK() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	checkBalance := int64(100500)
	s.mockRepo.On("GetBalanceByAccountID", mock.Anything, testAccount.ID).Return(checkBalance, true, nil)

	balance, found, err := s.operationUsecase.GetBalanceByAccountID(context.Background(), testAccount.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), checkBalance, balance)
	assert.Equal(s.T(), true, found)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *OperationTestSuite) TestGetBalanceByAccountID__Err() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	checkBalance := int64(0)
	s.mockRepo.On("GetBalanceByAccountID", mock.Anything, testAccount.ID).Return(checkBalance, false, e.ErrInternal)

	balance, found, err := s.operationUsecase.GetBalanceByAccountID(context.Background(), testAccount.ID)
	assert.ErrorIs(s.T(), err, e.ErrInternal)
	assert.Equal(s.T(), checkBalance, balance)
	assert.Equal(s.T(), false, found)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *OperationTestSuite) TestSaveOperationIncrease__OK() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	checkBalance := int64(100500)
	operation := domain.NewOperation(domain.OperationTypeIncrease, testAccount.ID, checkBalance, domain.OperationSourceTypeDeposit, nil)

	s.mockRepo.On("Create", mock.Anything, operation).Return(checkBalance, nil)

	balance, err := s.operationUsecase.SaveOperation(context.Background(), operation)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), checkBalance, balance)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *OperationTestSuite) TestSaveOperationIncrease__Err__СoncurrentExec() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	checkBalance := int64(100500)
	operation := domain.NewOperation(domain.OperationTypeIncrease, testAccount.ID, checkBalance, domain.OperationSourceTypeDeposit, nil)

	s.mockRepo.On("Create", mock.Anything, operation).Return(checkBalance, e.ErrTxСoncurrentExec)

	balance, err := s.operationUsecase.SaveOperation(context.Background(), operation)
	assert.ErrorIs(s.T(), err, e.ErrTxСoncurrentExec)
	assert.Equal(s.T(), int64(0), balance)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *OperationTestSuite) TestSaveOperationDecrease__OK() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	checkBalance := int64(100500)
	operation := domain.NewOperation(domain.OperationTypeDecrease, testAccount.ID, checkBalance, domain.OperationSourceTypeDeposit, nil)

	s.mockRepo.On("Create", mock.Anything, operation).Return(checkBalance, nil)

	balance, err := s.operationUsecase.SaveOperation(context.Background(), operation)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), checkBalance, balance)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *OperationTestSuite) TestSaveOperationDecrease__Err__СoncurrentExec() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	checkBalance := int64(100500)
	operation := domain.NewOperation(domain.OperationTypeDecrease, testAccount.ID, checkBalance, domain.OperationSourceTypeDeposit, nil)

	s.mockRepo.On("Create", mock.Anything, operation).Return(checkBalance, e.ErrTxСoncurrentExec)

	balance, err := s.operationUsecase.SaveOperation(context.Background(), operation)
	assert.ErrorIs(s.T(), err, e.ErrTxСoncurrentExec)
	assert.Equal(s.T(), int64(0), balance)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *OperationTestSuite) TestSaveOperationDecrease__Err__OperationNotEnoughFunds() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	checkBalance := int64(100500)
	operation := domain.NewOperation(domain.OperationTypeDecrease, testAccount.ID, checkBalance, domain.OperationSourceTypeDeposit, nil)

	s.mockRepo.On("Create", mock.Anything, operation).Return(int64(-100), nil)

	balance, err := s.operationUsecase.SaveOperation(context.Background(), operation)
	assert.ErrorIs(s.T(), err, usecase.ErrOperationNotEnoughFunds)
	assert.Equal(s.T(), int64(0), balance)

	s.mockRepo.AssertExpectations(s.T())
}
