package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// connectMaxAttempts and connectBackoff control the retry policy when dialing
// RabbitMQ at startup. With these defaults the producer will spend up to ~30s
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

type EventProducer interface {
	PublishOrderCreated(orderID string, userID uint) error
	Close() error
}

type eventProducer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewEventProducer(url string) (EventProducer, error) {
	var (
		conn *amqp.Connection
		err  error
	)

	for attempt := 1; attempt <= connectMaxAttempts; attempt++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("[Producer] RabbitMQ connect attempt %d/%d failed: %v", attempt, connectMaxAttempts, err)
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
		_ = ch.Close()
		_ = conn.Close()
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

func (p *eventProducer) Close() error {
	if p == nil {
		return nil
	}
	if p.ch != nil {
		if err := p.ch.Close(); err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}
	}
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}
	return nil
}
