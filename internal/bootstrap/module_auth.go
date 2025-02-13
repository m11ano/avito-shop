package bootstrap

import (
	"github.com/m11ano/avito-shop/internal/usecase"
	"go.uber.org/fx"
)

var AuthModule = fx.Module(
	"auth_module",
	fx.Provide(
		fx.Annotate(usecase.NewAuthInpl, fx.As(new(usecase.Auth))),
	),
)
