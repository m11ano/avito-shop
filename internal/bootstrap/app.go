package bootstrap

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m11ano/avito-shop/internal/config"
	"github.com/m11ano/avito-shop/internal/migrations"
	"go.uber.org/fx"
)

var App = fx.Options(
	fx.Provide(NewLogger),
	fx.Provide(NewPgxv5),
	fx.Invoke(func(lc fx.Lifecycle, shutdowner fx.Shutdowner, logger *slog.Logger, config config.Config, dbpool *pgxpool.Pool) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				err := Pgxv5TestConnection(context.Background(), dbpool, logger, config.DB.MaxAttempt, config.DB.AttemptSleepSeconds)
				if err != nil {
					return err
				}

				err = migrations.RunMigrations(ctx, dbpool, logger)
				if err != nil {
					return err
				}

				return nil
			},
		})
	}),
)
