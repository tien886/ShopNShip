package repository

import (
	"github.com/google/uuid"
	"github.com/tien886/ShopNShip/order-service/internal/model"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(order *model.Order) error
	FindByID(id uuid.UUID) (*model.Order, error)
	FindByUserID(userID uint) ([]model.Order, error)
	UpdateStatus(id uuid.UUID, status model.OrderStatus) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) FindByID(id uuid.UUID) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Items").First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByUserID(userID uint) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.Preload("Items").Where("user_id = ?", userID).Find(&orders).Error
	return orders, err
}

func (r *orderRepository) UpdateStatus(id uuid.UUID, status model.OrderStatus) error {
	return r.db.Model(&model.Order{}).Where("id = ?", id).Update("status", status).Error
}
