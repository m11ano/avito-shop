package bootstrap

import (
	"github.com/m11ano/avito-shop/internal/repository"
	"github.com/m11ano/avito-shop/internal/usecase"
	"go.uber.org/fx"
)

var OperationModule = fx.Module(
	"operation_module",
	fx.Provide(
		fx.Private,
		fx.Annotate(repository.NewOperation, fx.As(new(usecase.OperationRepository))),
	),
	fx.Provide(
		fx.Annotate(usecase.NewOperationInpl, fx.As(new(usecase.Operation))),
	),
)
