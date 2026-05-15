package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	StatusPending   OrderStatus = "PENDING"
	StatusPaid      OrderStatus = "PAID"
	StatusCancelled OrderStatus = "CANCELLED"
	StatusCompleted OrderStatus = "COMPLETED"
)

type Order struct {
	ID         uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     uint        `gorm:"not null" json:"user_id"`
	Status     OrderStatus `gorm:"default:PENDING" json:"status"`
	TotalPrice float64     `gorm:"type:decimal(10,2);not null" json:"total_price"`
	Items      []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID   uuid.UUID `gorm:"type:uuid;not null" json:"order_id"`
	ItemName  string    `gorm:"not null" json:"item_name"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	Price     float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	CreatedAt time.Time `json:"created_at"`
}
