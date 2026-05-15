package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderEvent struct {
	Event     string    `json:"event"`
	OrderID   string    `json:"order_id"`
	UserID    uint      `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type EventProducer interface {
	PublishOrderCreated(orderID string, userID uint) error
}

type eventProducer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewEventProducer(url string) (EventProducer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare exchange
	err = ch.ExchangeDeclare(
		"order.events", // name
		"topic",        // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare an exchange: %w", err)
	}

	return &eventProducer{conn: conn, ch: ch}, nil
}

func (p *eventProducer) PublishOrderCreated(orderID string, userID uint) error {
	event := OrderEvent{
		Event:     "OrderCreated",
		OrderID:   orderID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(
		context.Background(),
		"order.events",  // exchange
		"order.created", // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
