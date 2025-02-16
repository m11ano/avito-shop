package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m11ano/avito-shop/internal/infra/config"
	"github.com/m11ano/avito-shop/internal/infra/db/migrations"
	"github.com/m11ano/avito-shop/internal/infra/db/txmngr"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var App = fx.Options(
	// Инфраструктура
	fx.Provide(NewLogger),
	fx.WithLogger(func(config config.Config) fxevent.Logger {
		if !config.App.UseLogger {
			return fxevent.NopLogger
		}
		return &fxevent.ConsoleLogger{
			W: os.Stdout,
		}
	}),
	fx.Provide(NewPgxv5),
	fx.Provide(func(config config.Config, logger *slog.Logger) *fiber.App {
		fiberApp := NewHTTPFiber(HTTPConfig{
			UnderProxy: config.HTTP.UnderProxy,
			UseTraceID: true,
			UseLogger:  true,
		}, logger)
		return fiberApp
	}),
	fx.Provide(func(pgxpool *pgxpool.Pool) (*manager.Manager, *trmpgx.CtxGetter) {
		return txmngr.New(pgxpool)
	}),
	// Бизнес логика
	AccountModule,
	OperationModule,
	AuthModule,
	ShopItemModule,
	ShopPurchaseModule,
	CoinTransferModule,
	// Delivery
	DeliveryHTTP,
	// Start && Stop invoke
	fx.Invoke(func(lc fx.Lifecycle, shutdowner fx.Shutdowner, logger *slog.Logger, config config.Config, dbpool *pgxpool.Pool, fiberApp *fiber.App) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				err := Pgxv5TestConnection(ctx, dbpool, logger, config.DB.MaxAttempt, config.DB.AttemptSleepSeconds)
				if err != nil {
					return err
				}
				logger.Info("Postgress connected")

				err = migrations.RunMigrations(ctx, dbpool, config, logger)
				if err != nil {
					return err
				}

				if config.HTTP.Port > 0 {
					go func() {
						if err := fiberApp.Listen(fmt.Sprintf(":%d", config.HTTP.Port)); err != nil {
							logger.Error("failed to start fiber", slog.Any("error", err), slog.Any("trackeback", string(debug.Stack())))
							err := shutdowner.Shutdown()
							if err != nil {
								logger.Error("failed to shutdown", slog.Any("error", err), slog.Any("trackeback", string(debug.Stack())))
							}
						}
					}()
				}

				return nil
			},
			OnStop: func(_ context.Context) error {
				if config.HTTP.Port > 0 {
					logger.Info("stopping HTTP Fiber")
					err := fiberApp.ShutdownWithTimeout(time.Duration(config.HTTP.StopTimeout) * time.Second)
					if err != nil {
						logger.Error("failed to stop fiber", slog.Any("error", err), slog.Any("trackeback", string(debug.Stack())))
					}
				}

				logger.Info("stopping Postgress")
				dbpool.Close()

				return nil
			},
		})
	}),
)
