package bootstrap

import (
	"github.com/m11ano/avito-shop/internal/repository"
	"github.com/m11ano/avito-shop/internal/usecase"
	"go.uber.org/fx"
)

var CoinTransferModule = fx.Module(
	"coin_transfer_module",
	fx.Provide(
		fx.Private,
		fx.Annotate(repository.NewCoinTransfer, fx.As(new(usecase.CoinTransferRepository))),
	),
	fx.Provide(
		fx.Annotate(usecase.NewCoinTransferInpl, fx.As(new(usecase.CoinTransfer))),
	),
)
