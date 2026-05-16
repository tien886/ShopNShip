package service

import (
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/tien886/ShopNShip/delivery-service/internal/model"
	"github.com/tien886/ShopNShip/delivery-service/internal/repository"
)

var (
	ErrDeliveryNotFound = errors.New("delivery not found")
)

type DeliveryService interface {
	CreateDeliveryFromOrder(orderID string, userID uint) error
	GetDelivery(id uuid.UUID) (*model.Delivery, error)
	GetUserDeliveries(userID uint) ([]model.Delivery, error)
	GetAllDeliveries() ([]model.Delivery, error)
	UpdateStatus(id uuid.UUID, status string) error
}

type deliveryService struct {
	repo repository.DeliveryRepository
}

func NewDeliveryService(repo repository.DeliveryRepository) DeliveryService {
	return &deliveryService{repo: repo}
}

func (s *deliveryService) CreateDeliveryFromOrder(orderID string, userID uint) error {
	log.Printf("[FLOW][ORDER->DELIVERY][STEP 6/7][DELIVERY-SVC] create delivery request received order_id=%s user_id=%d", orderID, userID)

	parsedOrderID, err := uuid.Parse(orderID)
	if err != nil {
		return err
	}

	existing, _ := s.repo.FindByOrderID(parsedOrderID)
	if existing != nil {
		log.Printf("[FLOW][ORDER->DELIVERY][STEP 7/7][DELIVERY-SVC] delivery already exists order_id=%s delivery_id=%s (idempotent)", orderID, existing.ID)
		return nil
	}

	delivery := &model.Delivery{
		ID:      uuid.New(),
		OrderID: parsedOrderID,
		UserID:  userID,
		Status:  model.DeliveryStatusPending,
	}

	if err := s.repo.Create(delivery); err != nil {
		return err
	}

	log.Printf("[FLOW][ORDER->DELIVERY][STEP 7/7][DELIVERY-SVC] delivery persisted delivery_id=%s order_id=%s user_id=%d status=%s", delivery.ID, delivery.OrderID, delivery.UserID, delivery.Status)
	return nil
}

func (s *deliveryService) GetDelivery(id uuid.UUID) (*model.Delivery, error) {
	delivery, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrDeliveryNotFound
	}
	return delivery, nil
}

func (s *deliveryService) GetUserDeliveries(userID uint) ([]model.Delivery, error) {
	return s.repo.FindByUserID(userID)
}

func (s *deliveryService) GetAllDeliveries() ([]model.Delivery, error) {
	return s.repo.FindAll()
}

func (s *deliveryService) UpdateStatus(id uuid.UUID, status string) error {
	delivery, err := s.repo.FindByID(id)
	if err != nil {
		return ErrDeliveryNotFound
	}

	newStatus := model.DeliveryStatus(status)
	validTransitions := map[model.DeliveryStatus][]model.DeliveryStatus{
		model.DeliveryStatusPending:   {model.DeliveryStatusAssigned, model.DeliveryStatusCancelled},
		model.DeliveryStatusAssigned:  {model.DeliveryStatusInTransit, model.DeliveryStatusCancelled},
		model.DeliveryStatusInTransit: {model.DeliveryStatusDelivered, model.DeliveryStatusCancelled},
	}

	allowed, ok := validTransitions[delivery.Status]
	if !ok {
		return errors.New("cannot update status from current state")
	}

	valid := false
	for _, s := range allowed {
		if newStatus == s {
			valid = true
			break
		}
	}
	if !valid {
		return errors.New("invalid status transition")
	}

	return s.repo.UpdateStatus(id, newStatus)
}
