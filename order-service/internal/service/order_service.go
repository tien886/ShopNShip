package service

import (
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/tien886/ShopNShip/order-service/internal/dto"
	"github.com/tien886/ShopNShip/order-service/internal/event"
	"github.com/tien886/ShopNShip/order-service/internal/model"
	"github.com/tien886/ShopNShip/order-service/internal/repository"
)

var (
	ErrOrderNotFound = errors.New("order not found")
)

type OrderService interface {
	CreateOrder(userID uint, req dto.CreateOrderRequest) (*model.Order, error)
	GetOrder(id uuid.UUID) (*model.Order, error)
	GetUserOrders(userID uint) ([]model.Order, error)
	UpdateOrderStatus(id uuid.UUID, status string) error
}

type orderService struct {
	repo     repository.OrderRepository
	producer event.EventProducer
}

func NewOrderService(repo repository.OrderRepository, producer event.EventProducer) OrderService {
	return &orderService{
		repo:     repo,
		producer: producer,
	}
}

func (s *orderService) CreateOrder(userID uint, req dto.CreateOrderRequest) (*model.Order, error) {
	log.Printf("[FLOW][ORDER->DELIVERY][STEP 1/7][ORDER-SVC] create order request received user_id=%d item_count=%d", userID, len(req.Items))

	var totalPrice float64
	items := make([]model.OrderItem, len(req.Items))
	orderID := uuid.New()

	for i, item := range req.Items {
		totalPrice += item.Price * float64(item.Quantity)
		items[i] = model.OrderItem{
			ID:       uuid.New(),
			OrderID:  orderID,
			ItemName: item.ItemName,
			Quantity: item.Quantity,
			Price:    item.Price,
		}
	}

	order := &model.Order{
		ID:         orderID,
		UserID:     userID,
		Status:     model.StatusPending,
		TotalPrice: totalPrice,
		Items:      items,
	}

	if err := s.repo.Create(order); err != nil {
		return nil, err
	}
	log.Printf("[FLOW][ORDER->DELIVERY][STEP 2/7][ORDER-SVC] order persisted order_id=%s user_id=%d status=%s total_price=%.2f", order.ID, order.UserID, order.Status, order.TotalPrice)

	// Publish event
	if s.producer != nil {
		log.Printf("[FLOW][ORDER->DELIVERY][STEP 3/7][ORDER-SVC] publishing OrderCreated event order_id=%s user_id=%d", order.ID, order.UserID)
		if err := s.producer.PublishOrderCreated(order.ID.String(), order.UserID); err != nil {
			log.Printf("[FLOW][ORDER->DELIVERY][STEP 4/7][ORDER-SVC] publish failed order_id=%s user_id=%d err=%v", order.ID, order.UserID, err)
		} else {
			log.Printf("[FLOW][ORDER->DELIVERY][STEP 4/7][ORDER-SVC] publish success order_id=%s user_id=%d exchange=order.events routing_key=order.created", order.ID, order.UserID)
		}
	} else {
		log.Printf("[FLOW][ORDER->DELIVERY][STEP 3/7][ORDER-SVC] producer not initialized, skipping event order_id=%s user_id=%d", order.ID, order.UserID)
	}

	return order, nil
}

func (s *orderService) GetOrder(id uuid.UUID) (*model.Order, error) {
	order, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (s *orderService) GetUserOrders(userID uint) ([]model.Order, error) {
	return s.repo.FindByUserID(userID)
}

func (s *orderService) UpdateOrderStatus(id uuid.UUID, status string) error {
	return s.repo.UpdateStatus(id, model.OrderStatus(status))
}
