package models

import "github.com/google/uuid"

type OrderID uuid.UUID

func (v OrderID) String() string {
	return uuid.UUID(v).String()
}

type UserID uint64

type SKUID uint64

type WarehouseID uint64

type DeliveryVariantID uint64
