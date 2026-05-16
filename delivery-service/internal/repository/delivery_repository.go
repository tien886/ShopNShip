package repository

import (
	"github.com/google/uuid"
	"github.com/tien886/ShopNShip/delivery-service/internal/model"
	"gorm.io/gorm"
)

type DeliveryRepository interface {
	Create(delivery *model.Delivery) error
	FindByID(id uuid.UUID) (*model.Delivery, error)
	FindByOrderID(orderID uuid.UUID) (*model.Delivery, error)
	FindByUserID(userID uint) ([]model.Delivery, error)
	FindAll() ([]model.Delivery, error)
	UpdateStatus(id uuid.UUID, status model.DeliveryStatus) error
}

type deliveryRepository struct {
	db *gorm.DB
}

func NewDeliveryRepository(db *gorm.DB) DeliveryRepository {
	return &deliveryRepository{db: db}
}

func (r *deliveryRepository) Create(delivery *model.Delivery) error {
	return r.db.Create(delivery).Error
}

func (r *deliveryRepository) FindByID(id uuid.UUID) (*model.Delivery, error) {
	var delivery model.Delivery
	if err := r.db.First(&delivery, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (r *deliveryRepository) FindByOrderID(orderID uuid.UUID) (*model.Delivery, error) {
	var delivery model.Delivery
	if err := r.db.Where("order_id = ?", orderID).First(&delivery).Error; err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (r *deliveryRepository) FindByUserID(userID uint) ([]model.Delivery, error) {
	var deliveries []model.Delivery
	if err := r.db.Where("user_id = ?", userID).Find(&deliveries).Error; err != nil {
		return nil, err
	}
	return deliveries, nil
}

func (r *deliveryRepository) FindAll() ([]model.Delivery, error) {
	var deliveries []model.Delivery
	if err := r.db.Find(&deliveries).Error; err != nil {
		return nil, err
	}
	return deliveries, nil
}

func (r *deliveryRepository) UpdateStatus(id uuid.UUID, status model.DeliveryStatus) error {
	return r.db.Model(&model.Delivery{}).Where("id = ?", id).Update("status", status).Error
}
