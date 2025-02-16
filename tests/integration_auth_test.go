package integration_test

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

type AuthResponse struct {
	Token string `json:"token"`
}

func (s *IntegrationTestSuite) parseJWTAuthToken(tokenStr string) (jwt.MapClaims, bool, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(_ *jwt.Token) (interface{}, error) {
		return []byte(s.config.Auth.JWTSecretKey), nil
	})

	return claims, token.Valid, err
}

func (s *IntegrationTestSuite) TestSignInAndSignUp() {
	testcases := []Case{
		// первая авторизация
		{
			name:       "first auth",
			timeoustMs: 50,
			request: Request{
				method: http.MethodPost,
				path:   "/api/auth",
				body: map[string]interface{}{
					"username": "test",
					"password": "test",
				},
			},
			expectStatusCode: http.StatusOK,
			parseType:        CaseParseTypeJSON,
			parseResponse:    &AuthResponse{},
			respCheck: RespCheck{
				func(resp any) {
					respParsed, ok := resp.(*AuthResponse)
					assert.True(s.T(), ok, "Response should be of type *AuthResponse")

					claims, valid, err := s.parseJWTAuthToken(respParsed.Token)
					assert.NoError(s.T(), err)
					assert.True(s.T(), valid)

					accountID, ok := claims["accountID"]
					if !ok {
						s.T().Fatalf("accountID not found in claims")
					}

					var dbAccountID uuid.UUID
					err = s.pgxpool.QueryRow(context.Background(), "SELECT account_id FROM account WHERE account_id = $1", accountID).Scan(&dbAccountID)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), accountID, dbAccountID.String())

					var balance int64
					err = s.pgxpool.QueryRow(context.Background(), "SELECT balance FROM operation_balance WHERE account_id = $1", accountID).Scan(&balance)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), s.config.Auth.NewAccountAmount, balance)

					err = s.pgxpool.QueryRow(context.Background(), "SELECT COALESCE(SUM(amount), 0) as balance FROM operation WHERE account_id = $1", accountID).Scan(&balance)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), s.config.Auth.NewAccountAmount, balance)
				},
			},
		},
		// авторизация с неверным паролем
		{
			name:       "auth with incorrect password",
			timeoustMs: 50,
			request: Request{
				method: http.MethodPost,
				path:   "/api/auth",
				body: map[string]interface{}{
					"username": "test",
					"password": "test_incorrect",
				},
			},
			expectStatusCode: http.StatusUnauthorized,
		},
	}

	for _, testcase := range testcases {
		s.execTestCase(testcase)
	}
}
