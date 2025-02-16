package usecase_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/internal/app"
	"github.com/m11ano/avito-shop/internal/config"
	"github.com/m11ano/avito-shop/internal/db/txmngr"
	"github.com/m11ano/avito-shop/internal/domain"
	"github.com/m11ano/avito-shop/internal/usecase"
	"github.com/m11ano/avito-shop/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

type ShopPurchaseTestSuite struct {
	suite.Suite
	mockPgxPool          *mocks.PgxPool
	shopPurchaseUsecase  usecase.ShopPurchase
	app                  *fx.App
	mockRepo             *mocks.ShopPurchaseRepository
	mockAccountUsecase   *mocks.Account
	mockOperationUsecase *mocks.Operation
	mockShopItemUsecase  *mocks.ShopItem
}

func (s *ShopPurchaseTestSuite) SetupTest() {
	s.mockPgxPool = mocks.NewPgxPoolMockForTxManager()
	s.mockRepo = new(mocks.ShopPurchaseRepository)
	s.mockAccountUsecase = new(mocks.Account)
	s.mockOperationUsecase = new(mocks.Operation)
	s.mockShopItemUsecase = new(mocks.ShopItem)

	s.app = fx.New(
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
		fx.Provide(func() config.Config { return config.LoadConfig("../../../config.yml") }),
		fx.Provide(func() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }),
		fx.Provide(txmngr.NewProvider(s.mockPgxPool)),
		fx.Provide(func() usecase.ShopPurchaseRepository { return s.mockRepo }),
		fx.Provide(func() usecase.Account { return s.mockAccountUsecase }),
		fx.Provide(func() usecase.Operation { return s.mockOperationUsecase }),
		fx.Provide(func() usecase.ShopItem { return s.mockShopItemUsecase }),
		fx.Provide(fx.Annotate(usecase.NewShopPurchaseInpl, fx.As(new(usecase.ShopPurchase)))),
		fx.Populate(&s.shopPurchaseUsecase),
	)

	err := s.app.Start(context.Background())
	assert.NoError(s.T(), err)
}

func (s *ShopPurchaseTestSuite) TearDownTest() {
	err := s.app.Stop(context.Background())
	assert.NoError(s.T(), err)
}

func TestShopPurchaseSuiteRun(t *testing.T) {
	suite.Run(t, new(ShopPurchaseTestSuite))
}

func (s *ShopPurchaseTestSuite) TestGetInventory__OK() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	testShopItem := &domain.ShopItem{
		ID:    uuid.New(),
		Name:  "test",
		Price: int64(100),
	}

	checkInventory := []usecase.ShopPurchaseGetInventoryItem{
		{
			ShopItem: testShopItem,
			Quantity: int64(10),
		},
	}

	checkInventoryRepo := []usecase.ShopPurchaseRepositoryAggrInventoryItem{
		{
			ShopItemID: testShopItem.ID,
			Quantity:   int64(10),
		},
	}
	s.mockRepo.On("AggrInventoryByAccountID", mock.Anything, testAccount.ID).Return(checkInventoryRepo, nil)
	s.mockShopItemUsecase.On("GetItemsByIDs", mock.Anything, []uuid.UUID{testShopItem.ID}).Return(map[uuid.UUID]domain.ShopItem{testShopItem.ID: *testShopItem}, nil)

	history, err := s.shopPurchaseUsecase.GetInventory(context.Background(), testAccount.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), checkInventory, history)

	s.mockRepo.AssertExpectations(s.T())
	s.mockShopItemUsecase.AssertExpectations(s.T())
}

func (s *ShopPurchaseTestSuite) TestGetInventory__Err_Repo() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	s.mockRepo.On("AggrInventoryByAccountID", mock.Anything, testAccount.ID).Return(nil, app.ErrInternal)

	history, err := s.shopPurchaseUsecase.GetInventory(context.Background(), testAccount.ID)
	assert.ErrorIs(s.T(), err, app.ErrInternal)
	assert.Nil(s.T(), history)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *ShopPurchaseTestSuite) TestGetInventory__Err_ShopItemUsecase() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	testShopItem := domain.NewShopItem("test", 100)

	checkInventoryRepo := []usecase.ShopPurchaseRepositoryAggrInventoryItem{
		{
			ShopItemID: testShopItem.ID,
			Quantity:   int64(10),
		},
	}
	s.mockRepo.On("AggrInventoryByAccountID", mock.Anything, testAccount.ID).Return(checkInventoryRepo, nil)
	s.mockShopItemUsecase.On("GetItemsByIDs", mock.Anything, []uuid.UUID{testShopItem.ID}).Return(nil, app.ErrInternal)

	history, err := s.shopPurchaseUsecase.GetInventory(context.Background(), testAccount.ID)
	assert.ErrorIs(s.T(), err, app.ErrInternal)
	assert.Nil(s.T(), history)

	s.mockRepo.AssertExpectations(s.T())
	s.mockShopItemUsecase.AssertExpectations(s.T())
}

func (s *ShopPurchaseTestSuite) TestMakePurchase__OK() {
	testAccount, err := domain.NewAccount("test_owner", "test")
	assert.NoError(s.T(), err)

	testShopItem := domain.NewShopItem("test", 100)

	identityKey := uuid.New()

	quantity := int64(1)

	s.mockRepo.On("FindIdentity", mock.Anything, identityKey).Return(false, nil)
	s.mockShopItemUsecase.On("GetItemByName", mock.Anything, testShopItem.Name).Return(testShopItem, nil)
	s.mockOperationUsecase.On("SaveOperation", mock.Anything, mock.AnythingOfType("*domain.Operation")).Return(int64(0), nil).Once()
	s.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.ShopPurchase")).Return(nil).Once()

	purchase, err := s.shopPurchaseUsecase.MakePurchase(context.Background(), testShopItem.Name, testAccount.ID, quantity, &identityKey)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), testAccount.ID, purchase.AccountID)
	assert.Equal(s.T(), testShopItem.ID, purchase.ItemID)
	assert.Equal(s.T(), quantity, purchase.Quantity)

	s.mockRepo.AssertExpectations(s.T())
	s.mockShopItemUsecase.AssertExpectations(s.T())
	s.mockOperationUsecase.AssertExpectations(s.T())
}

func (s *ShopPurchaseTestSuite) TestMakePurchase__Err_Identity() {
	testAccount, err := domain.NewAccount("test_owner", "test")
	assert.NoError(s.T(), err)

	testShopItem := domain.NewShopItem("test", 100)

	identityKey := uuid.New()

	quantity := int64(1)

	s.mockRepo.On("FindIdentity", mock.Anything, identityKey).Return(true, nil)

	purchase, err := s.shopPurchaseUsecase.MakePurchase(context.Background(), testShopItem.Name, testAccount.ID, quantity, &identityKey)
	assert.ErrorIs(s.T(), err, app.ErrConflict)
	assert.Nil(s.T(), purchase)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *ShopPurchaseTestSuite) TestMakePurchase__Err_IdentityInternal() {
	testAccount, err := domain.NewAccount("test_owner", "test")
	assert.NoError(s.T(), err)

	testShopItem := domain.NewShopItem("test", 100)

	identityKey := uuid.New()

	quantity := int64(1)

	s.mockRepo.On("FindIdentity", mock.Anything, identityKey).Return(false, app.ErrInternal)

	purchase, err := s.shopPurchaseUsecase.MakePurchase(context.Background(), testShopItem.Name, testAccount.ID, quantity, &identityKey)
	assert.ErrorIs(s.T(), err, app.ErrInternal)
	assert.Nil(s.T(), purchase)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *ShopPurchaseTestSuite) TestMakePurchase__Err_ShopItemNotFound() {
	testAccount, err := domain.NewAccount("test_owner", "test")
	assert.NoError(s.T(), err)

	testShopItem := domain.NewShopItem("test", 100)

	identityKey := uuid.New()

	quantity := int64(1)

	s.mockRepo.On("FindIdentity", mock.Anything, identityKey).Return(false, nil)
	s.mockShopItemUsecase.On("GetItemByName", mock.Anything, testShopItem.Name).Return(nil, app.ErrNotFound)

	purchase, err := s.shopPurchaseUsecase.MakePurchase(context.Background(), testShopItem.Name, testAccount.ID, quantity, &identityKey)
	assert.ErrorIs(s.T(), err, app.ErrNotFound)
	assert.Nil(s.T(), purchase)

	s.mockRepo.AssertExpectations(s.T())
	s.mockShopItemUsecase.AssertExpectations(s.T())
}

func (s *ShopPurchaseTestSuite) TestMakePurchase__Err_FundsNotEnough() {
	testAccount, err := domain.NewAccount("test_owner", "test")
	assert.NoError(s.T(), err)

	testShopItem := domain.NewShopItem("test", 100)

	identityKey := uuid.New()

	quantity := int64(1)

	s.mockRepo.On("FindIdentity", mock.Anything, identityKey).Return(false, nil)
	s.mockShopItemUsecase.On("GetItemByName", mock.Anything, testShopItem.Name).Return(testShopItem, nil)
	s.mockOperationUsecase.On("SaveOperation", mock.Anything, mock.AnythingOfType("*domain.Operation")).Return(int64(0), usecase.ErrOperationNotEnoughFunds).Once()

	purchase, err := s.shopPurchaseUsecase.MakePurchase(context.Background(), testShopItem.Name, testAccount.ID, quantity, &identityKey)
	assert.ErrorIs(s.T(), err, usecase.ErrOperationNotEnoughFunds)
	assert.Nil(s.T(), purchase)

	s.mockRepo.AssertExpectations(s.T())
	s.mockShopItemUsecase.AssertExpectations(s.T())
	s.mockOperationUsecase.AssertExpectations(s.T())
}

func (s *ShopPurchaseTestSuite) TestMakePurchase__Err_TargetSaveErr() {
	testAccount, err := domain.NewAccount("test_owner", "test")
	assert.NoError(s.T(), err)

	testShopItem := domain.NewShopItem("test", 100)

	identityKey := uuid.New()

	quantity := int64(1)

	someErr := errors.New("some error")

	s.mockRepo.On("FindIdentity", mock.Anything, identityKey).Return(false, nil)
	s.mockShopItemUsecase.On("GetItemByName", mock.Anything, testShopItem.Name).Return(testShopItem, nil)
	s.mockOperationUsecase.On("SaveOperation", mock.Anything, mock.AnythingOfType("*domain.Operation")).Return(int64(0), nil).Once()
	s.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.ShopPurchase")).Return(someErr).Once()

	purchase, err := s.shopPurchaseUsecase.MakePurchase(context.Background(), testShopItem.Name, testAccount.ID, quantity, &identityKey)
	assert.ErrorIs(s.T(), err, someErr)
	assert.Nil(s.T(), purchase)

	s.mockRepo.AssertExpectations(s.T())
	s.mockShopItemUsecase.AssertExpectations(s.T())
	s.mockOperationUsecase.AssertExpectations(s.T())
}

/*


func (s *ShopPurchaseTestSuite) TestMakeTransferByUsername__Err_TargetSaveErr() {
	ownerAccount, err := domain.NewAccount("test_owner", "test")
	assert.NoError(s.T(), err)

	targetAccount, err := domain.NewAccount("test_target", "test")
	assert.NoError(s.T(), err)

	identityKey := uuid.New()

	amount := int64(100500)

	s.mockRepo.On("FindIdentity", mock.Anything, identityKey).Return(false, nil)
	s.mockAccountUsecase.On("GetItemByUsername", mock.Anything, targetAccount.Username).Return(targetAccount, nil)
	s.mockOperationUsecase.On("SaveOperation", mock.Anything, mock.AnythingOfType("*domain.Operation")).Return(int64(0), nil).Once()
	s.mockOperationUsecase.On("SaveOperation", mock.Anything, mock.AnythingOfType("*domain.Operation")).Return(int64(0), app.ErrInternal).Once()

	ownerShopPurchase, targetShopPurchase, err := s.shopPurchaseUsecase.MakeTransferByUsername(context.Background(), targetAccount.Username, ownerAccount.ID, amount, &identityKey)
	assert.ErrorIs(s.T(), err, app.ErrInternal)
	assert.Nil(s.T(), ownerShopPurchase)
	assert.Nil(s.T(), targetShopPurchase)

	s.mockRepo.AssertExpectations(s.T())
	s.mockAccountUsecase.AssertExpectations(s.T())
	s.mockOperationUsecase.AssertExpectations(s.T())
}

func (s *ShopPurchaseTestSuite) TestMakeTransferByUsername__Err_OwnerRepoSave() {
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
	s.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.ShopPurchase")).Return(someErr).Once()

	ownerShopPurchase, targetShopPurchase, err := s.shopPurchaseUsecase.MakeTransferByUsername(context.Background(), targetAccount.Username, ownerAccount.ID, amount, &identityKey)
	assert.ErrorIs(s.T(), err, someErr)
	assert.Nil(s.T(), ownerShopPurchase)
	assert.Nil(s.T(), targetShopPurchase)

	s.mockRepo.AssertExpectations(s.T())
	s.mockAccountUsecase.AssertExpectations(s.T())
	s.mockOperationUsecase.AssertExpectations(s.T())
}

func (s *ShopPurchaseTestSuite) TestMakeTransferByUsername__Err_TargetRepoSave() {
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
	s.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.ShopPurchase")).Return(nil).Once()
	s.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.ShopPurchase")).Return(someErr).Once()

	ownerShopPurchase, targetShopPurchase, err := s.shopPurchaseUsecase.MakeTransferByUsername(context.Background(), targetAccount.Username, ownerAccount.ID, amount, &identityKey)
	assert.ErrorIs(s.T(), err, someErr)
	assert.Nil(s.T(), ownerShopPurchase)
	assert.Nil(s.T(), targetShopPurchase)

	s.mockRepo.AssertExpectations(s.T())
	s.mockAccountUsecase.AssertExpectations(s.T())
	s.mockOperationUsecase.AssertExpectations(s.T())
}
*/
