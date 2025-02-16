package usecase_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
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

type CoinTransferTestSuite struct {
	suite.Suite
	mockPgxPool          *mocks.PgxPool
	coinTransferUsecase  usecase.CoinTransfer
	app                  *fx.App
	mockRepo             *mocks.CoinTransferRepository
	mockAccountUsecase   *mocks.Account
	mockOperationUsecase *mocks.Operation
}

func (s *CoinTransferTestSuite) SetupTest() {
	s.mockPgxPool = mocks.NewPgxPoolMockForTxManager()
	s.mockRepo = new(mocks.CoinTransferRepository)
	s.mockAccountUsecase = new(mocks.Account)
	s.mockOperationUsecase = new(mocks.Operation)

	s.app = fx.New(
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
		fx.Provide(func() config.Config { return config.LoadConfig("../../../config.yml") }),
		fx.Provide(func() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }),
		fx.Provide(txmngr.NewProvider(s.mockPgxPool)),
		fx.Provide(func() usecase.CoinTransferRepository { return s.mockRepo }),
		fx.Provide(func() usecase.Account { return s.mockAccountUsecase }),
		fx.Provide(func() usecase.Operation { return s.mockOperationUsecase }),
		fx.Provide(fx.Annotate(usecase.NewCoinTransferInpl, fx.As(new(usecase.CoinTransfer)))),
		fx.Populate(&s.coinTransferUsecase),
	)

	err := s.app.Start(context.Background())
	assert.NoError(s.T(), err)
}

func (s *CoinTransferTestSuite) TearDownTest() {
	err := s.app.Stop(context.Background())
	assert.NoError(s.T(), err)
}

func TestCoinTransferSuiteRun(t *testing.T) {
	suite.Run(t, new(CoinTransferTestSuite))
}

func (s *CoinTransferTestSuite) TestGetAggrCoinHistory__OK() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	checkHistory := []usecase.CoinTransferGetAggrHistoryItem{
		{
			Account: testAccount,
			Amount:  int64(100500),
		},
	}

	checkHistoryRepo := []usecase.CoinTransferRepositoryAggrHistoryItem{
		{
			AccountID: testAccount.ID,
			Amount:    int64(100500),
		},
	}
	s.mockRepo.On("GetAggrCoinHistoryByAccountID", mock.Anything, testAccount.ID, mock.AnythingOfType("domain.CoinTransferType")).Return(checkHistoryRepo, nil)
	s.mockAccountUsecase.On("GetItemsByIDs", mock.Anything, []uuid.UUID{testAccount.ID}).Return(map[uuid.UUID]domain.Account{testAccount.ID: *testAccount}, nil)

	history, err := s.coinTransferUsecase.GetAggrCoinHistory(context.Background(), testAccount.ID, domain.CoinTransferTypeReciving)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), checkHistory, history)

	s.mockRepo.AssertExpectations(s.T())
	s.mockAccountUsecase.AssertExpectations(s.T())
}

func (s *CoinTransferTestSuite) TestGetAggrCoinHistory__Err_Repo() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	s.mockRepo.On("GetAggrCoinHistoryByAccountID", mock.Anything, testAccount.ID, mock.AnythingOfType("domain.CoinTransferType")).Return(nil, e.ErrInternal)

	history, err := s.coinTransferUsecase.GetAggrCoinHistory(context.Background(), testAccount.ID, domain.CoinTransferTypeReciving)
	assert.ErrorIs(s.T(), err, e.ErrInternal)
	assert.Nil(s.T(), history)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *CoinTransferTestSuite) TestGetAggrCoinHistory__Err_AccountUsecase() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	checkHistoryRepo := []usecase.CoinTransferRepositoryAggrHistoryItem{
		{
			AccountID: testAccount.ID,
			Amount:    int64(100500),
		},
	}
	s.mockRepo.On("GetAggrCoinHistoryByAccountID", mock.Anything, testAccount.ID, mock.AnythingOfType("domain.CoinTransferType")).Return(checkHistoryRepo, nil)
	s.mockAccountUsecase.On("GetItemsByIDs", mock.Anything, []uuid.UUID{testAccount.ID}).Return(nil, e.ErrInternal)

	history, err := s.coinTransferUsecase.GetAggrCoinHistory(context.Background(), testAccount.ID, domain.CoinTransferTypeReciving)
	assert.ErrorIs(s.T(), err, e.ErrInternal)
	assert.Nil(s.T(), history)

	s.mockRepo.AssertExpectations(s.T())
	s.mockAccountUsecase.AssertExpectations(s.T())
}

func (s *CoinTransferTestSuite) TestMakeTransferByUsername__OK() {
	ownerAccount, err := domain.NewAccount("test_owner", "test")
	assert.NoError(s.T(), err)

	targetAccount, err := domain.NewAccount("test_target", "test")
	assert.NoError(s.T(), err)

	identityKey := uuid.New()

	amount := int64(100500)

	s.mockRepo.On("FindIdentity", mock.Anything, identityKey).Return(false, nil)
	s.mockAccountUsecase.On("GetItemByUsername", mock.Anything, targetAccount.Username).Return(targetAccount, nil)
	s.mockOperationUsecase.On("SaveOperation", mock.Anything, mock.AnythingOfType("*domain.Operation")).Return(int64(0), nil).Once()
	s.mockOperationUsecase.On("SaveOperation", mock.Anything, mock.AnythingOfType("*domain.Operation")).Return(int64(0), nil).Once()
	s.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.CoinTransfer")).Return(nil).Once()
	s.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.CoinTransfer")).Return(nil).Once()

	ownerCoinTransfer, targetCoinTransfer, err := s.coinTransferUsecase.MakeTransferByUsername(context.Background(), targetAccount.Username, ownerAccount.ID, amount, &identityKey)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), ownerAccount.ID, ownerCoinTransfer.OwnerAccountID)
	assert.Equal(s.T(), targetAccount.ID, ownerCoinTransfer.CounterpartyAccountID)
	assert.Equal(s.T(), amount, ownerCoinTransfer.Amount)
	assert.Equal(s.T(), domain.CoinTransferTypeSending, ownerCoinTransfer.Type)

	assert.Equal(s.T(), targetAccount.ID, targetCoinTransfer.OwnerAccountID)
	assert.Equal(s.T(), ownerAccount.ID, targetCoinTransfer.CounterpartyAccountID)
	assert.Equal(s.T(), amount, targetCoinTransfer.Amount)
	assert.Equal(s.T(), domain.CoinTransferTypeReciving, targetCoinTransfer.Type)

	s.mockRepo.AssertExpectations(s.T())
	s.mockAccountUsecase.AssertExpectations(s.T())
	s.mockOperationUsecase.AssertExpectations(s.T())
}

func (s *CoinTransferTestSuite) TestMakeTransferByUsername__Err_ZeroAmount() {
	ownerAccount, err := domain.NewAccount("test_owner", "test")
	assert.NoError(s.T(), err)

	targetAccount, err := domain.NewAccount("test_target", "test")
	assert.NoError(s.T(), err)

	identityKey := uuid.New()

	amount := int64(0)
	ownerCoinTransfer, targetCoinTransfer, err := s.coinTransferUsecase.MakeTransferByUsername(context.Background(), targetAccount.Username, ownerAccount.ID, amount, &identityKey)
	assert.ErrorIs(s.T(), err, usecase.ErrCoinTransferAmountMustBeGreaterThanZero)
	assert.Nil(s.T(), ownerCoinTransfer)
	assert.Nil(s.T(), targetCoinTransfer)
}

func (s *CoinTransferTestSuite) TestMakeTransferByUsername__Err_Identity() {
	ownerAccount, err := domain.NewAccount("test_owner", "test")
	assert.NoError(s.T(), err)

	targetAccount, err := domain.NewAccount("test_target", "test")
	assert.NoError(s.T(), err)

	identityKey := uuid.New()

	amount := int64(100500)

	s.mockRepo.On("FindIdentity", mock.Anything, identityKey).Return(true, nil)
	ownerCoinTransfer, targetCoinTransfer, err := s.coinTransferUsecase.MakeTransferByUsername(context.Background(), targetAccount.Username, ownerAccount.ID, amount, &identityKey)
	assert.ErrorIs(s.T(), err, e.ErrConflict)
	assert.Nil(s.T(), ownerCoinTransfer)
	assert.Nil(s.T(), targetCoinTransfer)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *CoinTransferTestSuite) TestMakeTransferByUsername__Err_IdentityInternal() {
	ownerAccount, err := domain.NewAccount("test_owner", "test")
	assert.NoError(s.T(), err)

	targetAccount, err := domain.NewAccount("test_target", "test")
	assert.NoError(s.T(), err)

	identityKey := uuid.New()

	amount := int64(100500)

	s.mockRepo.On("FindIdentity", mock.Anything, identityKey).Return(false, e.ErrInternal)
	ownerCoinTransfer, targetCoinTransfer, err := s.coinTransferUsecase.MakeTransferByUsername(context.Background(), targetAccount.Username, ownerAccount.ID, amount, &identityKey)
	assert.ErrorIs(s.T(), err, e.ErrInternal)
	assert.Nil(s.T(), ownerCoinTransfer)
	assert.Nil(s.T(), targetCoinTransfer)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *CoinTransferTestSuite) TestMakeTransferByUsername__Err_TargetNotFound() {
	ownerAccount, err := domain.NewAccount("test_owner", "test")
	assert.NoError(s.T(), err)

	targetAccount, err := domain.NewAccount("test_target", "test")
	assert.NoError(s.T(), err)

	identityKey := uuid.New()

	amount := int64(100500)

	s.mockRepo.On("FindIdentity", mock.Anything, identityKey).Return(false, nil)
	s.mockAccountUsecase.On("GetItemByUsername", mock.Anything, targetAccount.Username).Return(nil, e.ErrNotFound)

	ownerCoinTransfer, targetCoinTransfer, err := s.coinTransferUsecase.MakeTransferByUsername(context.Background(), targetAccount.Username, ownerAccount.ID, amount, &identityKey)
	assert.ErrorIs(s.T(), err, e.ErrNotFound)
	assert.Nil(s.T(), ownerCoinTransfer)
	assert.Nil(s.T(), targetCoinTransfer)

	s.mockRepo.AssertExpectations(s.T())
	s.mockAccountUsecase.AssertExpectations(s.T())
}

func (s *CoinTransferTestSuite) TestMakeTransferByUsername__Err_SameOwnerAndTarget() {
	ownerAccount, err := domain.NewAccount("test_owner", "test")
	assert.NoError(s.T(), err)

	targetAccount := ownerAccount

	identityKey := uuid.New()

	amount := int64(100500)

	s.mockRepo.On("FindIdentity", mock.Anything, identityKey).Return(false, nil)
	s.mockAccountUsecase.On("GetItemByUsername", mock.Anything, targetAccount.Username).Return(targetAccount, nil)

	ownerCoinTransfer, targetCoinTransfer, err := s.coinTransferUsecase.MakeTransferByUsername(context.Background(), targetAccount.Username, ownerAccount.ID, amount, &identityKey)
	assert.ErrorIs(s.T(), err, e.ErrConflict)
	assert.Nil(s.T(), ownerCoinTransfer)
	assert.Nil(s.T(), targetCoinTransfer)

	s.mockRepo.AssertExpectations(s.T())
	s.mockAccountUsecase.AssertExpectations(s.T())
}

func (s *CoinTransferTestSuite) TestMakeTransferByUsername__Err_FundsNotEnough() {
	ownerAccount, err := domain.NewAccount("test_owner", "test")
	assert.NoError(s.T(), err)

	targetAccount, err := domain.NewAccount("test_target", "test")
	assert.NoError(s.T(), err)

	identityKey := uuid.New()

	amount := int64(100500)

	s.mockRepo.On("FindIdentity", mock.Anything, identityKey).Return(false, nil)
	s.mockAccountUsecase.On("GetItemByUsername", mock.Anything, targetAccount.Username).Return(targetAccount, nil)
	s.mockOperationUsecase.On("SaveOperation", mock.Anything, mock.AnythingOfType("*domain.Operation")).Return(int64(0), usecase.ErrOperationNotEnoughFunds)

	ownerCoinTransfer, targetCoinTransfer, err := s.coinTransferUsecase.MakeTransferByUsername(context.Background(), targetAccount.Username, ownerAccount.ID, amount, &identityKey)
	assert.ErrorIs(s.T(), err, usecase.ErrOperationNotEnoughFunds)
	assert.Nil(s.T(), ownerCoinTransfer)
	assert.Nil(s.T(), targetCoinTransfer)

	s.mockRepo.AssertExpectations(s.T())
	s.mockAccountUsecase.AssertExpectations(s.T())
	s.mockOperationUsecase.AssertExpectations(s.T())
}

func (s *CoinTransferTestSuite) TestMakeTransferByUsername__Err_TargetSaveErr() {
	ownerAccount, err := domain.NewAccount("test_owner", "test")
	assert.NoError(s.T(), err)

	targetAccount, err := domain.NewAccount("test_target", "test")
	assert.NoError(s.T(), err)

	identityKey := uuid.New()

	amount := int64(100500)

	s.mockRepo.On("FindIdentity", mock.Anything, identityKey).Return(false, nil)
	s.mockAccountUsecase.On("GetItemByUsername", mock.Anything, targetAccount.Username).Return(targetAccount, nil)
	s.mockOperationUsecase.On("SaveOperation", mock.Anything, mock.AnythingOfType("*domain.Operation")).Return(int64(0), nil).Once()
	s.mockOperationUsecase.On("SaveOperation", mock.Anything, mock.AnythingOfType("*domain.Operation")).Return(int64(0), e.ErrInternal).Once()

	ownerCoinTransfer, targetCoinTransfer, err := s.coinTransferUsecase.MakeTransferByUsername(context.Background(), targetAccount.Username, ownerAccount.ID, amount, &identityKey)
	assert.ErrorIs(s.T(), err, e.ErrInternal)
	assert.Nil(s.T(), ownerCoinTransfer)
	assert.Nil(s.T(), targetCoinTransfer)

	s.mockRepo.AssertExpectations(s.T())
	s.mockAccountUsecase.AssertExpectations(s.T())
	s.mockOperationUsecase.AssertExpectations(s.T())
}

func (s *CoinTransferTestSuite) TestMakeTransferByUsername__Err_OwnerRepoSave() {
	ownerAccount, err := domain.NewAccount("test_owner", "test")
	assert.NoError(s.T(), err)

	targetAccount, err := domain.NewAccount("test_target", "test")
	assert.NoError(s.T(), err)

	identityKey := uuid.New()

	amount := int64(100500)

	someErr := errors.New("some error")

	s.mockRepo.On("FindIdentity", mock.Anything, identityKey).Return(false, nil)
	s.mockAccountUsecase.On("GetItemByUsername", mock.Anything, targetAccount.Username).Return(targetAccount, nil)
	s.mockOperationUsecase.On("SaveOperation", mock.Anything, mock.AnythingOfType("*domain.Operation")).Return(int64(0), nil).Once()
	s.mockOperationUsecase.On("SaveOperation", mock.Anything, mock.AnythingOfType("*domain.Operation")).Return(int64(0), nil).Once()
	s.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.CoinTransfer")).Return(someErr).Once()

	ownerCoinTransfer, targetCoinTransfer, err := s.coinTransferUsecase.MakeTransferByUsername(context.Background(), targetAccount.Username, ownerAccount.ID, amount, &identityKey)
	assert.ErrorIs(s.T(), err, someErr)
	assert.Nil(s.T(), ownerCoinTransfer)
	assert.Nil(s.T(), targetCoinTransfer)

	s.mockRepo.AssertExpectations(s.T())
	s.mockAccountUsecase.AssertExpectations(s.T())
	s.mockOperationUsecase.AssertExpectations(s.T())
}

func (s *CoinTransferTestSuite) TestMakeTransferByUsername__Err_TargetRepoSave() {
	ownerAccount, err := domain.NewAccount("test_owner", "test")
	assert.NoError(s.T(), err)

	targetAccount, err := domain.NewAccount("test_target", "test")
	assert.NoError(s.T(), err)

	identityKey := uuid.New()

	amount := int64(100500)

	someErr := errors.New("some error")

	s.mockRepo.On("FindIdentity", mock.Anything, identityKey).Return(false, nil)
	s.mockAccountUsecase.On("GetItemByUsername", mock.Anything, targetAccount.Username).Return(targetAccount, nil)
	s.mockOperationUsecase.On("SaveOperation", mock.Anything, mock.AnythingOfType("*domain.Operation")).Return(int64(0), nil).Once()
	s.mockOperationUsecase.On("SaveOperation", mock.Anything, mock.AnythingOfType("*domain.Operation")).Return(int64(0), nil).Once()
	s.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.CoinTransfer")).Return(nil).Once()
	s.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.CoinTransfer")).Return(someErr).Once()

	ownerCoinTransfer, targetCoinTransfer, err := s.coinTransferUsecase.MakeTransferByUsername(context.Background(), targetAccount.Username, ownerAccount.ID, amount, &identityKey)
	assert.ErrorIs(s.T(), err, someErr)
	assert.Nil(s.T(), ownerCoinTransfer)
	assert.Nil(s.T(), targetCoinTransfer)

	s.mockRepo.AssertExpectations(s.T())
	s.mockAccountUsecase.AssertExpectations(s.T())
	s.mockOperationUsecase.AssertExpectations(s.T())
}
