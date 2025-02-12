package bootstrap

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m11ano/avito-shop/internal/config"
	"go.uber.org/fx"
)

func NewPgxv5(lc fx.Lifecycle, logger *slog.Logger, config config.Config) *pgxpool.Pool {

	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, config.DB.URI)
	if err != nil {
		panic("unable to create pgxv5 connection pool")
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			logger.Info("stopping Postgress")
			dbpool.Close()
			return nil
		},
	})

	return dbpool
}

func Pgxv5TestConnection(ctx context.Context, dbpool *pgxpool.Pool, logger *slog.Logger, maxAttempt int, attemptSleepSeconds int) error {
	attemp := 1
	var err error
	for attemp <= maxAttempt {
		err = dbpool.Ping(ctx)
		if err != nil {
			logger.Info("failed to connect to Postgress", slog.Int("attemp", attemp))
			time.Sleep(time.Duration(attemptSleepSeconds) * time.Second)
			attemp++
			continue
		}
		break
	}
	if err != nil {
		return err
	}

	return nil
}
