# Order–Delivery Communication Workflow via RabbitMQ

This document summarizes how `order-service` and `delivery-service` communicate asynchronously through RabbitMQ.

## Components

- **Producer:** `order-service`
- **Consumer:** `delivery-service`
- **Broker:** RabbitMQ
- **Exchange:** `order.events` (type: `topic`, durable)
- **Queue:** `delivery.order.created` (durable)
- **Binding:** `delivery.order.created` <- `order.created` on `order.events`

## End-to-End Flow

1. Client creates an order via `POST /orders`.
2. `order-service` validates request, calculates total price, and persists order with `pending` status.
3. `order-service` publishes an `OrderCreated` event to exchange `order.events` using routing key `order.created`.
4. RabbitMQ routes the message to queue `delivery.order.created`.
5. `delivery-service` consumer receives the message (manual ack mode).
6. Consumer unmarshals payload and calls `CreateDeliveryFromOrder(order_id, user_id)`.
7. `delivery-service` creates a delivery record with `pending` status (idempotent by `order_id`) and then sends ACK.

## Event Payload

```json
{
  "event": "OrderCreated",
  "order_id": "uuid",
  "user_id": 123,
  "created_at": "2026-05-17T14:51:18Z"
}
```

## Reliability Notes

- Both producer and consumer retry RabbitMQ connection on startup (up to 10 attempts, 3s backoff).
- Consumer uses **manual acknowledgments**:
  - `Ack` after successful delivery creation.
  - `Nack(requeue=true)` on business processing failure (retry).
  - `Nack(requeue=false)` on invalid/unmarshalable payload (drop).
- Delivery creation is idempotent: if a delivery already exists for `order_id`, service returns success without creating duplicates.

## Sequence (Simplified)

```text
Client -> Order Service: POST /orders
Order Service -> Order DB: INSERT order + items
Order Service -> RabbitMQ(order.events): Publish order.created
RabbitMQ -> delivery.order.created: Route message
Delivery Consumer -> Delivery Service: CreateDeliveryFromOrder
Delivery Service -> Delivery DB: INSERT delivery (if not exists)
Delivery Consumer -> RabbitMQ: ACK
```
