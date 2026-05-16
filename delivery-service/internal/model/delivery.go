package model

import (
	"time"

	"github.com/google/uuid"
)

type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "PENDING"
	DeliveryStatusAssigned  DeliveryStatus = "ASSIGNED"
	DeliveryStatusInTransit DeliveryStatus = "IN_TRANSIT"
	DeliveryStatusDelivered DeliveryStatus = "DELIVERED"
	DeliveryStatusCancelled DeliveryStatus = "CANCELLED"
)

type Delivery struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID   uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex" json:"order_id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	Status    DeliveryStatus `gorm:"default:PENDING" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
