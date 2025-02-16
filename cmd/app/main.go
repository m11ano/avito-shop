package main

import (
	"time"

	"github.com/m11ano/avito-shop/internal/bootstrap"
	"github.com/m11ano/avito-shop/internal/infra/config"
	"go.uber.org/fx"
)

func main() {
	cfg := config.LoadConfig("config.yml")

	app := fx.New(
		fx.Options(
			fx.StartTimeout(time.Second*time.Duration(cfg.App.StartTimeout)),
			fx.StopTimeout(time.Second*time.Duration(cfg.App.StopTimeout)),
		),
		fx.Provide(func() config.Config {
			return cfg
		}),
		bootstrap.App,
	)

	app.Run()
}
