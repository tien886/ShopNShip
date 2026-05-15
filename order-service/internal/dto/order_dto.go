package dto

import "github.com/google/uuid"

type CreateOrderRequest struct {
	Items []OrderItemDTO `json:"items" binding:"required,dive"`
}

type OrderItemDTO struct {
	ItemName string  `json:"item_name" binding:"required"`
	Quantity int     `json:"quantity" binding:"required,gt=0"`
	Price    float64 `json:"price" binding:"required,gt=0"`
}

type OrderResponse struct {
	ID         uuid.UUID      `json:"id"`
	UserID     uint           `json:"user_id"`
	Status     string         `json:"status"`
	TotalPrice float64        `json:"total_price"`
	Items      []OrderItemDTO `json:"items"`
	CreatedAt  string         `json:"created_at"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=PENDING PAID CANCELLED COMPLETED"`
}
