package bootstrap

import (
	"github.com/m11ano/avito-shop/internal/repository"
	"github.com/m11ano/avito-shop/internal/usecase"
	"go.uber.org/fx"
)

var ShopItemModule = fx.Module(
	"shop_item_module",
	fx.Provide(
		fx.Private,
		fx.Annotate(repository.NewShopItem, fx.As(new(usecase.ShopItemRepository))),
	),
	fx.Provide(
		fx.Annotate(usecase.NewShopItemInpl, fx.As(new(usecase.ShopItem))),
	),
)
