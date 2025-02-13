package bootstrap

import (
	"github.com/m11ano/avito-shop/internal/repository"
	"github.com/m11ano/avito-shop/internal/usecase"
	"go.uber.org/fx"
)

var ShopPurchaseModule = fx.Module(
	"shop_purchase_module",
	fx.Provide(
		fx.Private,
		fx.Annotate(repository.NewShopPurchase, fx.As(new(usecase.ShopPurchaseRepository))),
	),
	fx.Provide(
		fx.Annotate(usecase.NewShopPurchaseInpl, fx.As(new(usecase.ShopPurchase))),
	),
)
