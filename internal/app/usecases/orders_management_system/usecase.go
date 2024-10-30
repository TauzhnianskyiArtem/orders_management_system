package orders_management_system

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/moguchev/microservices_courcse/orders_management_system/internal/app/models"
)

var (
	ErrReserveStocks = errors.New("failed to reserve stock")
)

type UsecaseInterface interface {
	CreateOrder(ctx context.Context, userID models.UserID, info CreateOrderInfo) (*models.Order, error)
}

//go:generate mockery --name=WarehouseManagementSystem --filename=warehouse_management_system_mock.go --disable-version-string
//go:generate mockery --name=OrdersStorage --filename=orders_storage_mock.go --disable-version-string

type (
	WarehouseManagementSystem interface {
		ReserveStocks(ctx context.Context, userID models.UserID, items []models.Item) error
	}

	OrdersStorage interface {
		CreateOrder(ctx context.Context, order *models.Order) error
		CreateOutboxMessage(ctx context.Context, order *models.Order) error
	}

	TransactionManager interface {
		RunReadCommitted(ctx context.Context, accessMode pgx.TxAccessMode, f func(ctx context.Context) error) error
	}
)

type Deps struct {
	TransactionManager
	WarehouseManagementSystem
	OrdersStorage
}

type usecase struct {
	Deps
}

func NewUsecase(d Deps) UsecaseInterface {
	return &usecase{
		Deps: d,
	}
}
