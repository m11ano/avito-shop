package txmngr

import (
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/avito-tech/go-transaction-manager/trm/v2/settings"
	"github.com/jackc/pgx/v5"
	"github.com/m11ano/avito-shop/internal/infra/db"
)

func New(dbpool db.PgxPool) (*manager.Manager, *trmpgx.CtxGetter) {
	txOptions := pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead, // Устанавливаем уровень изоляции RepeatableRead
	}

	settingsOpts := settings.Must()

	trmpgxSettings := trmpgx.MustSettings(settingsOpts, trmpgx.WithTxOptions(txOptions))

	txFactory := trmpgx.NewFactory(dbpool)

	txManager := manager.Must(txFactory, manager.WithSettings(trmpgxSettings))

	return txManager, trmpgx.DefaultCtxGetter
}
