package usecase_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

type AuthTestSuite struct {
	suite.Suite
	config               config.Config
	mockPgxPool          *mocks.PgxPool
	authUsecase          usecase.Auth
	app                  *fx.App
	mockAccountUsecase   *mocks.Account
	mockOperationUsecase *mocks.Operation
}

func (s *AuthTestSuite) SetupTest() {
	s.config = config.LoadConfig("../../../config.yml")
	s.mockPgxPool = mocks.NewPgxPoolMockForTxManager()
	s.mockAccountUsecase = new(mocks.Account)
	s.mockOperationUsecase = new(mocks.Operation)

	s.app = fx.New(
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
		fx.Provide(func() config.Config { return s.config }),
		fx.Provide(func() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }),
		fx.Provide(txmngr.NewProvider(s.mockPgxPool)),
		fx.Provide(func() usecase.Account { return s.mockAccountUsecase }),
		fx.Provide(func() usecase.Operation { return s.mockOperationUsecase }),
		fx.Provide(fx.Annotate(usecase.NewAuthInpl, fx.As(new(usecase.Auth)))),
		fx.Populate(&s.authUsecase),
	)

	err := s.app.Start(context.Background())
	assert.NoError(s.T(), err)
}

func (s *AuthTestSuite) TearDownTest() {
	err := s.app.Stop(context.Background())
	assert.NoError(s.T(), err)
}

func TestAuthSuiteRun(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

func (s *AuthTestSuite) TestSignIn__OK() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	s.mockAccountUsecase.On("GetItemByUsername", mock.Anything, testAccount.Username).Return(testAccount, nil)

	jwtToken, err := s.authUsecase.SignInOrSignUp(context.Background(), "test", "test")
	assert.NoError(s.T(), err)
	_, err = jwt.Parse(jwtToken, func(_ *jwt.Token) (interface{}, error) {
		return []byte(s.config.Auth.JWTSecretKey), nil
	})
	assert.NoError(s.T(), err)

	s.mockAccountUsecase.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestSignIn__InvalidPassword() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	s.mockAccountUsecase.On("GetItemByUsername", mock.Anything, testAccount.Username).Return(testAccount, nil)

	jwtToken, err := s.authUsecase.SignInOrSignUp(context.Background(), "test", "test_invalid")
	assert.ErrorIs(s.T(), err, e.ErrUnauthorized)
	assert.Empty(s.T(), jwtToken)

	s.mockAccountUsecase.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestSignIn__ErrInternal() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	s.mockAccountUsecase.On("GetItemByUsername", mock.Anything, testAccount.Username).Return(nil, e.ErrInternal)

	jwtToken, err := s.authUsecase.SignInOrSignUp(context.Background(), "test", "test")
	assert.ErrorIs(s.T(), err, e.ErrInternal)
	assert.Empty(s.T(), jwtToken)

	s.mockAccountUsecase.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestSignIn__ErrUnknown() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)
	unknownErr := errors.New("unknown")
	s.mockAccountUsecase.On("GetItemByUsername", mock.Anything, testAccount.Username).Return(nil, unknownErr)

	jwtToken, err := s.authUsecase.SignInOrSignUp(context.Background(), "test", "test")
	assert.ErrorIs(s.T(), err, unknownErr)
	assert.Empty(s.T(), jwtToken)

	s.mockAccountUsecase.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestSignUp__OK() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	s.mockAccountUsecase.On("GetItemByUsername", mock.Anything, testAccount.Username).Return(nil, e.ErrNotFound)
	s.mockAccountUsecase.On("Create", mock.Anything, mock.AnythingOfType("*domain.Account")).Return(nil)
	s.mockOperationUsecase.On("SaveOperation", mock.Anything, mock.AnythingOfType("*domain.Operation")).Return(s.config.Auth.NewAccountAmount, nil)

	jwtToken, err := s.authUsecase.SignInOrSignUp(context.Background(), "test", "test")
	assert.NoError(s.T(), err)
	_, err = jwt.Parse(jwtToken, func(_ *jwt.Token) (interface{}, error) {
		return []byte(s.config.Auth.JWTSecretKey), nil
	})
	assert.NoError(s.T(), err)

	s.mockAccountUsecase.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestSignUp__ErrInternal() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	s.mockAccountUsecase.On("GetItemByUsername", mock.Anything, testAccount.Username).Return(nil, e.ErrNotFound)
	s.mockAccountUsecase.On("Create", mock.Anything, mock.AnythingOfType("*domain.Account")).Return(e.ErrInternal)

	jwtToken, err := s.authUsecase.SignInOrSignUp(context.Background(), "test", "test")
	assert.ErrorIs(s.T(), err, e.ErrInternal)
	assert.Empty(s.T(), jwtToken)

	s.mockAccountUsecase.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestSignUp__ErrTxСoncurrentExec() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	s.mockAccountUsecase.On("GetItemByUsername", mock.Anything, testAccount.Username).Return(nil, e.ErrNotFound)
	s.mockAccountUsecase.On("Create", mock.Anything, mock.AnythingOfType("*domain.Account")).Return(nil)
	s.mockOperationUsecase.On("SaveOperation", mock.Anything, mock.AnythingOfType("*domain.Operation")).Return(int64(0), e.ErrTxСoncurrentExec)

	jwtToken, err := s.authUsecase.SignInOrSignUp(context.Background(), "test", "test")
	assert.ErrorIs(s.T(), err, e.ErrTxСoncurrentExec)
	assert.Empty(s.T(), jwtToken)

	s.mockAccountUsecase.AssertExpectations(s.T())
	s.mockOperationUsecase.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestAuthByJWTToken__OK() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"accountID": testAccount.ID.String(),
		"createdAt": strconv.FormatInt(time.Now().Unix(), 10),
	})

	tokenStr, err := token.SignedString([]byte(s.config.Auth.JWTSecretKey))
	assert.NoError(s.T(), err)

	accountID, err := s.authUsecase.AuthByJWTToken(context.Background(), tokenStr)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), testAccount.ID, *accountID)
}

func (s *AuthTestSuite) TestAuthByJWTToken__Err_Empty() {
	accountID, err := s.authUsecase.AuthByJWTToken(context.Background(), "")
	assert.ErrorIs(s.T(), err, e.ErrUnauthorized)
	assert.Nil(s.T(), accountID)
}

func (s *AuthTestSuite) TestAuthByJWTToken__Err_EmptyAccountID() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"createdAt": strconv.FormatInt(time.Now().Unix(), 10),
	})

	tokenStr, err := token.SignedString([]byte(s.config.Auth.JWTSecretKey))
	assert.NoError(s.T(), err)

	accountID, err := s.authUsecase.AuthByJWTToken(context.Background(), tokenStr)
	assert.ErrorIs(s.T(), err, e.ErrUnauthorized)
	assert.Nil(s.T(), accountID)
}

func (s *AuthTestSuite) TestAuthByJWTToken__Err_EmptyCreatedAt() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"accountID": testAccount.ID.String(),
	})

	tokenStr, err := token.SignedString([]byte(s.config.Auth.JWTSecretKey))
	assert.NoError(s.T(), err)

	accountID, err := s.authUsecase.AuthByJWTToken(context.Background(), tokenStr)
	assert.ErrorIs(s.T(), err, e.ErrUnauthorized)
	assert.Nil(s.T(), accountID)
}

func (s *AuthTestSuite) TestAuthByJWTToken__Err_OldCreatedAt() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"accountID": testAccount.ID.String(),
		"createdAt": strconv.FormatInt(time.Now().Add(time.Duration(-s.config.Auth.JWTTokenTTL-1)*time.Second).Unix(), 10),
	})

	tokenStr, err := token.SignedString([]byte(s.config.Auth.JWTSecretKey))
	assert.NoError(s.T(), err)

	accountID, err := s.authUsecase.AuthByJWTToken(context.Background(), tokenStr)
	assert.ErrorIs(s.T(), err, e.ErrUnauthorized)
	assert.Nil(s.T(), accountID)
}

func (s *AuthTestSuite) TestAuthByJWTToken__Err_InvalidCreatedAt() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"accountID": testAccount.ID.String(),
		"createdAt": "text",
	})

	tokenStr, err := token.SignedString([]byte(s.config.Auth.JWTSecretKey))
	assert.NoError(s.T(), err)

	accountID, err := s.authUsecase.AuthByJWTToken(context.Background(), tokenStr)
	assert.ErrorIs(s.T(), err, e.ErrUnauthorized)
	assert.Nil(s.T(), accountID)
}

func (s *AuthTestSuite) TestAuthByJWTToken__Err_InvalidAccountID() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"accountID": "invalid",
		"createdAt": strconv.FormatInt(time.Now().Unix(), 10),
	})

	tokenStr, err := token.SignedString([]byte(s.config.Auth.JWTSecretKey))
	assert.NoError(s.T(), err)

	accountID, err := s.authUsecase.AuthByJWTToken(context.Background(), tokenStr)
	assert.ErrorIs(s.T(), err, e.ErrUnauthorized)
	assert.Nil(s.T(), accountID)
}

func (s *AuthTestSuite) TestAuthByJWTToken__Err_InvalidJWT() {
	testAccount, err := domain.NewAccount("test", "test")
	assert.NoError(s.T(), err)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"accountID": testAccount.ID.String(),
		"createdAt": strconv.FormatInt(time.Now().Unix(), 10),
	})

	tokenStr, err := token.SignedString([]byte("invalid secret key"))
	assert.NoError(s.T(), err)

	accountID, err := s.authUsecase.AuthByJWTToken(context.Background(), tokenStr)
	assert.ErrorIs(s.T(), err, e.ErrUnauthorized)
	assert.Nil(s.T(), accountID)
}
