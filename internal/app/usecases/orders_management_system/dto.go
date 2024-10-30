package orders_management_system

import "github.com/moguchev/microservices_courcse/orders_management_system/internal/app/models"

type CreateOrderInfo struct {
	Items             []models.Item
	DeliveryOrderInfo models.DeliveryOrderInfo
}
