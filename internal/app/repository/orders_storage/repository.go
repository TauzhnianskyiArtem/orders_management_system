package orders_storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
	oms "github.com/moguchev/microservices_courcse/orders_management_system/internal/app/usecases/orders_management_system"
	"github.com/moguchev/microservices_courcse/orders_management_system/pkg/postgres"
	"github.com/moguchev/microservices_courcse/orders_management_system/pkg/transaction_manager"
)

var (
	_ oms.OrdersStorage = (*OrdersStorage)(nil)
)

type Connection interface {
	Execx(ctx context.Context, sqlizer postgres.Sqlizer) (pgconn.CommandTag, error)
}

type OrdersStorage struct {
	driver QueryEngineProvider
}

type QueryEngineProvider interface {
	GetQueryEngine(ctx context.Context) transaction_manager.QueryEngine
}

func New(driver QueryEngineProvider) *OrdersStorage {
	return &OrdersStorage{
		driver: driver,
	}
}

const (
	tableOrdersName = "orders"
)
