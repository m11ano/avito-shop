package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func RunMigrations(_ context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	if err := goose.SetDialect("postgres"); err != nil {
		logger.Error("goose: unable to set dialect", slog.Any("error", err))
		return err
	}
	goose.SetLogger(NewSlogLogger(logger))

	db := stdlib.OpenDBFromPool(pool)
	if err := goose.Up(db, "migrations"); err != nil {
		logger.Error("goose: unable to run migrations", slog.Any("error", err))
		return err
	}

	return nil
}
