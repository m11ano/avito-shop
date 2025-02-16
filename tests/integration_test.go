package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/m11ano/avito-shop/internal/bootstrap"
	"github.com/m11ano/avito-shop/internal/infra/config"
	"github.com/m11ano/avito-shop/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/fx"
)

type IntegrationTestSuite struct {
	suite.Suite
	config           config.Config
	app              *fx.App
	fiberApp         *fiber.App
	pgxpool          *pgxpool.Pool
	accountUsecase   usecase.Account
	operationUsecase usecase.Operation
	shopItemUsecase  usecase.ShopItem
	authUsecase      usecase.Auth
}

func (s *IntegrationTestSuite) SetupTest() {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Ошибка при создании контейнера: %v", err)
	}

	s.config = config.LoadConfig("../config.yml")
	s.config.DB.MigrationsPath = "../migrations"
	s.config.HTTP.Port = 0
	s.config.App.UseFxLogger = false
	s.config.App.UseLogger = false

	//nolint:errcheck
	host, _ := container.Host(ctx)
	//nolint:errcheck
	port, _ := container.MappedPort(ctx, "5432/tcp")
	s.config.DB.URI = fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())

	s.app = fx.New(
		fx.Options(
			fx.StartTimeout(time.Second*time.Duration(s.config.App.StartTimeout)),
			fx.StopTimeout(time.Second*time.Duration(s.config.App.StopTimeout)),
		),
		fx.Provide(func() config.Config {
			return s.config
		}),
		bootstrap.App,
		fx.Populate(&s.fiberApp, &s.pgxpool, &s.accountUsecase, &s.operationUsecase, &s.shopItemUsecase, &s.authUsecase),
	)

	err = s.app.Start(context.Background())
	assert.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) TearDownTest() {
	err := s.app.Stop(context.Background())
	assert.NoError(s.T(), err)
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

type Request struct {
	method  string
	path    string
	headers map[string]string
	body    map[string]interface{}
}

type RespCheck []func(any)

const CaseParseTypeJSON = "json"

const CaseParseTypeText = "string"

type Case struct {
	name             string
	request          Request
	timeoustMs       int
	expectStatusCode int
	parseType        string
	parseResponse    interface{}
	respCheck        RespCheck
}

func (s *IntegrationTestSuite) execTestCase(testcase Case) {
	s.T().Logf("[%s] case started", testcase.name)

	var body []byte
	if testcase.request.body != nil {
		var err error
		body, err = json.Marshal(testcase.request.body)
		assert.NoError(s.T(), err)
	}

	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(testcase.request.method, testcase.request.path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(testcase.request.method, testcase.request.path, nil)
	}

	for key, value := range testcase.request.headers {
		req.Header.Set(key, value)
	}

	resp, err := s.fiberApp.Test(req, testcase.timeoustMs)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	if !assert.Equal(s.T(), testcase.expectStatusCode, resp.StatusCode) {
		return
	}

	respBody, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	switch testcase.parseType {
	case CaseParseTypeJSON:
		err = json.Unmarshal(respBody, testcase.parseResponse)
		assert.NoError(s.T(), err)
	case CaseParseTypeText:
		testcase.parseResponse = string(respBody)
	}

	for _, check := range testcase.respCheck {
		check(testcase.parseResponse)
	}
}
