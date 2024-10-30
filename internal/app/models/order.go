package models

import (
	"time"
)

type Order struct {
	ID     OrderID
	UserID UserID
	Items  []Item
	DeliveryOrderInfo
}

type DeliveryOrderInfo struct {
	DeliveryVariantID DeliveryVariantID
	DeliveryDate      time.Time
}

type Item struct {
	SKU         SKU
	Quantity    uint32
	WarehouseID WarehouseID
}
