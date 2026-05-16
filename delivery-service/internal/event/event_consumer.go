package event

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tien886/ShopNShip/delivery-service/internal/service"
)

// connectMaxAttempts and connectBackoff control the retry policy when dialing
// RabbitMQ at startup. With these defaults the consumer will spend up to ~30s
// trying to connect, which gives the broker container time to come up.
const (
	connectMaxAttempts = 10
	connectBackoff     = 3 * time.Second
)

type OrderEvent struct {
	Event     string    `json:"event"`
	OrderID   string    `json:"order_id"`
	UserID    uint      `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type EventConsumer struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue amqp.Queue
	svc   service.DeliveryService
}

func NewEventConsumer(url string, svc service.DeliveryService) (*EventConsumer, error) {
	var (
		conn *amqp.Connection
		err  error
	)

	for attempt := 1; attempt <= connectMaxAttempts; attempt++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("[Consumer] RabbitMQ connect attempt %d/%d failed: %v", attempt, connectMaxAttempts, err)
		if attempt < connectMaxAttempts {
			time.Sleep(connectBackoff)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", connectMaxAttempts, err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		"order.events",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	q, err := ch.QueueDeclare(
		"delivery.order.created",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.QueueBind(
		q.Name,
		"order.created",
		"order.events",
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &EventConsumer{conn: conn, ch: ch, queue: q, svc: svc}, nil
}

func (c *EventConsumer) Start() error {
	msgs, err := c.ch.Consume(
		c.queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		log.Println("[Consumer] started, waiting for messages...")
		for d := range msgs {
			log.Printf("[FLOW][ORDER->DELIVERY][STEP 5/7][CONSUMER] message received delivery_tag=%d bytes=%d", d.DeliveryTag, len(d.Body))

			var event OrderEvent
			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Printf("[FLOW][ORDER->DELIVERY][STEP 5/7][CONSUMER] unmarshal failed delivery_tag=%d err=%v", d.DeliveryTag, err)
				d.Nack(false, false)
				continue
			}

			log.Printf("[FLOW][ORDER->DELIVERY][STEP 6/7][CONSUMER] dispatching to delivery service event=%s order_id=%s user_id=%d", event.Event, event.OrderID, event.UserID)

			if err := c.svc.CreateDeliveryFromOrder(event.OrderID, event.UserID); err != nil {
				log.Printf("[FLOW][ORDER->DELIVERY][STEP 6/7][CONSUMER] delivery creation failed order_id=%s err=%v requeue=true", event.OrderID, err)
				d.Nack(false, true)
				continue
			}

			d.Ack(false)
			log.Printf("[FLOW][ORDER->DELIVERY][STEP 7/7][CONSUMER] ack sent order_id=%s delivery_tag=%d", event.OrderID, d.DeliveryTag)
		}
		log.Println("[Consumer] message channel closed")
	}()

	return nil
}

func (c *EventConsumer) Shutdown() error {
	if c.ch != nil {
		if err := c.ch.Close(); err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}
	return nil
}
