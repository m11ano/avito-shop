package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m11ano/avito-shop/internal/config"
)

type pgxv5Tracer struct {
	logger *slog.Logger
}

func (t *pgxv5Tracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	t.logger.Info(fmt.Sprintf("Начало запроса: %s, args: %v", data.SQL, data.Args))
	return ctx
}

func (t *pgxv5Tracer) TraceQueryEnd(_ context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	t.logger.Info(fmt.Sprintf("Завершение запроса: %s, ошибка: %v", data.CommandTag, data.Err))
}

func NewPgxv5(config config.Config, logger *slog.Logger) *pgxpool.Pool {
	ctx := context.Background()
	pgxCfg, err := pgxpool.ParseConfig(config.DB.URI)
	if err != nil {
		panic("unable to parse db uri string")
	}

	if config.App.Mode == "dev" {
		pgxCfg.ConnConfig.Tracer = &pgxv5Tracer{logger: logger}
	}

	dbpool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		panic("unable to create pgxv5 connection pool")
	}

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
