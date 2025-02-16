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
	"github.com/m11ano/avito-shop/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

type ShopItemTestSuite struct {
	suite.Suite
	mockPgxPool     *mocks.PgxPool
	shopItemUsecase usecase.ShopItem
	app             *fx.App
	mockRepo        *mocks.ShopItemRepository
}

func (s *ShopItemTestSuite) SetupTest() {
	s.mockPgxPool = mocks.NewPgxPoolMockForTxManager()
	s.mockRepo = new(mocks.ShopItemRepository)

	s.app = fx.New(
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
		fx.Provide(func() config.Config { return config.LoadConfig("../../../config.yml") }),
		fx.Provide(func() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }),
		fx.Provide(txmngr.NewProvider(s.mockPgxPool)),
		fx.Provide(func() usecase.ShopItemRepository { return s.mockRepo }),
		fx.Provide(fx.Annotate(usecase.NewShopItemInpl, fx.As(new(usecase.ShopItem)))),
		fx.Populate(&s.shopItemUsecase),
	)

	err := s.app.Start(context.Background())
	assert.NoError(s.T(), err)
}

func (s *ShopItemTestSuite) TearDownTest() {
	err := s.app.Stop(context.Background())
	assert.NoError(s.T(), err)
}

func TestShopItemSuiteRun(t *testing.T) {
	suite.Run(t, new(ShopItemTestSuite))
}

func (s *ShopItemTestSuite) TestGetItemByID__OK() {
	testShopItem := domain.NewShopItem("test", 100)

	s.mockRepo.On("FindItemByID", mock.Anything, testShopItem.ID).Return(testShopItem, nil)

	shopItem, err := s.shopItemUsecase.GetItemByID(context.Background(), testShopItem.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), testShopItem, shopItem)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *ShopItemTestSuite) TestGetItemByID__NotFound() {
	shopItemID := uuid.New()
	s.mockRepo.On("FindItemByID", mock.Anything, shopItemID).Return(nil, app.ErrNotFound)

	shopItem, err := s.shopItemUsecase.GetItemByID(context.Background(), shopItemID)
	assert.ErrorIs(s.T(), err, app.ErrNotFound)
	assert.Nil(s.T(), shopItem)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *ShopItemTestSuite) TestGetItemByName__OK() {
	testShopItem := domain.NewShopItem("test", 100)

	s.mockRepo.On("FindItemByName", mock.Anything, testShopItem.Name).Return(testShopItem, nil)

	shopItem, err := s.shopItemUsecase.GetItemByName(context.Background(), "test")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), testShopItem, shopItem)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *ShopItemTestSuite) TestGetItemByName__NotFound() {
	s.mockRepo.On("FindItemByName", mock.Anything, "test").Return(nil, app.ErrNotFound)

	shopItem, err := s.shopItemUsecase.GetItemByName(context.Background(), "test")
	assert.ErrorIs(s.T(), err, app.ErrNotFound)
	assert.Nil(s.T(), shopItem)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *ShopItemTestSuite) TestGetItemByIDs__OK() {
	testShopItem1 := domain.NewShopItem("test1", 100)
	testShopItem2 := domain.NewShopItem("test2", 200)

	s.mockRepo.On("FindItemsByIDs", mock.Anything, []uuid.UUID{testShopItem1.ID, testShopItem2.ID}).Return(map[uuid.UUID]domain.ShopItem{testShopItem1.ID: *testShopItem1, testShopItem2.ID: *testShopItem2}, nil)

	shopItems, err := s.shopItemUsecase.GetItemsByIDs(context.Background(), []uuid.UUID{testShopItem1.ID, testShopItem2.ID})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), map[uuid.UUID]domain.ShopItem{testShopItem1.ID: *testShopItem1, testShopItem2.ID: *testShopItem2}, shopItems)

	s.mockRepo.AssertExpectations(s.T())
}
