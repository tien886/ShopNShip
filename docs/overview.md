# Delivery Management Microservice System

A learning-focused backend microservice project using Go, PostgreSQL, RabbitMQ, Docker, and REST APIs.

---

# 1. Project Goal

Build a small but realistic distributed backend system to learn:

- Microservice architecture
- REST API communication
- JWT authentication
- Event-driven architecture
- RabbitMQ messaging
- Docker containerization
- API Gateway
- Distributed system concepts
- Observability and monitoring

This project is intentionally small enough for learning but structured like a real-world backend system.

---

# 2. System Overview

## Services

```text
- API Gateway
- Authentication Service
- Order Service
- Delivery Service
```

---

# 3. High-Level Architecture

```text
                    Client
                       |
                 API Gateway
                       |
    ------------------------------------------------
    |                     |                        |
Auth Service        Order Service         Delivery Service
    |                     |                        |
PostgreSQL          PostgreSQL             PostgreSQL
                           |
                     RabbitMQ Broker
```

---

# 4. Tech Stack

| Category | Technology |
|---|---|
| Language | Go 1.22+ |
| Framework | Gin |
| Database | PostgreSQL |
| ORM | GORM |
| Authentication | JWT |
| Message Broker | RabbitMQ |
| Cache | Redis |
| API Gateway | NGINX |
| Containerization | Docker |
| API Documentation | Swagger |
| Logging | Zerolog |
| Monitoring | Prometheus |
| Dashboard | Grafana |
| Tracing | OpenTelemetry + Jaeger |

---

# 5. Main Business Flow

## Example User Flow

```text
1. User logs in
2. User creates an order
3. Order Service publishes "OrderCreated"
4. Delivery Service consumes event
5. Delivery task is created
6. Delivery status updates
```

---

# 6. Project Structure

```text
delivery-microservice-system/
│
├── gateway/
│
├── auth-service/
│
├── order-service/
│
├── delivery-service/
│
├── shared/
│   ├── events/
│   ├── middleware/
│   ├── logger/
│   ├── utils/
│   └── config/
│
├── infrastructure/
│   ├── nginx/
│   ├── prometheus/
│   ├── grafana/
│   └── rabbitmq/
│
├── docker-compose.yml
│
└── README.md
```

---

# 7. Internal Service Structure

Example:

```text
order-service/
│
├── cmd/
│
├── internal/
│   ├── handler/
│   ├── service/
│   ├── repository/
│   ├── middleware/
│   ├── dto/
│   ├── model/
│   ├── event/
│   ├── config/
│   └── validator/
│
├── migrations/
│
├── Dockerfile
│
├── .env
│
└── go.mod
```

---

# 8. Authentication Service

## Responsibilities

- Register user
- Login user
- Generate JWT
- Validate JWT
- Role management

---

## Database

### users

| Field | Type |
|---|---|
| id | UUID |
| email | VARCHAR |
| password_hash | VARCHAR |
| role | VARCHAR |
| created_at | TIMESTAMP |

---

## API Endpoints

| Method | Endpoint |
|---|---|
| POST | /register |
| POST | /login |
| GET | /profile |

---

## Concepts Learned

- JWT authentication
- Password hashing
- Authentication middleware
- Stateless auth

---

# 9. Order Service

## Responsibilities

- Create orders
- View orders
- Update order status
- Publish order events

---

## Database

### orders

| Field | Type |
|---|---|
| id | UUID |
| user_id | UUID |
| status | VARCHAR |
| total_price | DECIMAL |
| created_at | TIMESTAMP |

---

### order_items

| Field | Type |
|---|---|
| id | UUID |
| order_id | UUID |
| item_name | VARCHAR |
| quantity | INT |

---

## API Endpoints

| Method | Endpoint |
|---|---|
| POST | /orders |
| GET | /orders |
| GET | /orders/:id |

---

## Published Events

```text
OrderCreated
OrderCancelled
OrderCompleted
```

---

## Concepts Learned

- Service ownership
- Event publishing
- REST API design
- Database per service

---

# 10. Delivery Service

## Responsibilities

- Consume order events
- Create delivery task
- Update delivery status
- Assign delivery driver

---

## Database

### deliveries

| Field | Type |
|---|---|
| id | UUID |
| order_id | UUID |
| driver_name | VARCHAR |
| status | VARCHAR |
| created_at | TIMESTAMP |

---

## API Endpoints

| Method | Endpoint |
|---|---|
| GET | /deliveries |
| PATCH | /deliveries/:id/status |

---

## Consumed Events

```text
OrderCreated
```

---

## Published Events

```text
DeliveryAssigned
DeliveryCompleted
```

---

# 11. API Gateway

## Recommended Gateway

Use NGINX initially.

Later you may replace it with a Go gateway.

---

## Responsibilities

- Route requests
- JWT validation
- Rate limiting
- Reverse proxy
- Request logging

---

## Example Routing

| Route | Destination |
|---|---|
| /auth/* | Auth Service |
| /orders/* | Order Service |
| /delivery/* | Delivery Service |

---

# 12. Communication Design

# Synchronous Communication

Use REST APIs.

Example:

```text
Order Service
    -> Auth Service
    -> Validate JWT
```

---

# Asynchronous Communication

Use RabbitMQ.

Example:

```text
Order Service
    -> Publish OrderCreated

Delivery Service
    -> Consume OrderCreated
```

---

# 13. RabbitMQ Design

## Exchanges

```text
order.events
delivery.events
```

---

## Example Event Payload

```json
{
  "event": "OrderCreated",
  "order_id": "uuid",
  "user_id": "uuid",
  "created_at": "timestamp"
}
```

---

# 14. Docker Setup

## Containers

```text
- gateway
- auth-service
- order-service
- delivery-service
- postgres-auth
- postgres-order
- postgres-delivery
- rabbitmq
- redis
```

---

# 15. Database Strategy

Each service owns its own database.

```text
auth_db
order_db
delivery_db
```

Never directly query another service database.

---

# 16. Logging

Use Zerolog.

Log:

- request ID
- status code
- latency
- errors

---

# 17. Monitoring

## Prometheus

Track:

- HTTP requests
- latency
- RabbitMQ consumers
- error rates

---

## Grafana

Visualize:

- API metrics
- CPU usage
- DB performance
- queue status

---

# 18. Distributed Tracing

Use:

- OpenTelemetry
- Jaeger

Learn:

- trace requests across services
- debug distributed systems

---

# 19. Important Distributed System Concepts

## 1. Idempotency

RabbitMQ may redeliver events.

Consumers must safely process duplicate messages.

---

## 2. Retry Pattern

Handle:

- DB failure
- broker reconnect
- network issues

---

## 3. Graceful Shutdown

Learn:

- signal handling
- context cancellation

---

## 4. Eventual Consistency

Distributed systems rarely use global transactions.

---

# 20. Development Roadmap

# Phase 1 — Infrastructure Setup

## Tasks

- Setup project structure
- Setup Docker Compose
- Setup PostgreSQL
- Setup RabbitMQ
- Setup Redis
- Setup NGINX gateway

---

# Phase 2 — Authentication Service

## Tasks

- User model
- Register/Login APIs
- JWT middleware
- Password hashing
- Swagger docs

---

# Phase 3 — Order Service

## Tasks

- Order CRUD
- PostgreSQL integration
- JWT authentication
- Publish RabbitMQ events

---

# Phase 4 — Delivery Service

## Tasks

- RabbitMQ consumer
- Delivery task creation
- Delivery status update

---

# Phase 5 — Infrastructure Improvements

## Tasks

- Redis caching
- Centralized logging
- Retry mechanism
- Graceful shutdown

---

# Phase 6 — Observability

## Tasks

- Prometheus metrics
- Grafana dashboards
- OpenTelemetry tracing
- Jaeger setup

---

# 21. Recommended Learning Timeline

| Week | Focus |
|---|---|
| 1 | Infrastructure + Auth |
| 2 | Order Service |
| 3 | Delivery Service |
| 4 | RabbitMQ + Monitoring |

---

# 22. Recommended Go Libraries

| Purpose | Library |
|---|---|
| HTTP Framework | Gin |
| ORM | GORM |
| JWT | golang-jwt |
| Logging | Zerolog |
| Validation | validator |
| RabbitMQ | amqp091-go |
| Config | viper |
| Metrics | Prometheus client |
| Tracing | OpenTelemetry |

---

# 23. Future Improvements

After completing the core system:

- gRPC communication
- Custom Go API Gateway
- Service discovery
- Kubernetes deployment
- CI/CD pipeline
- Saga pattern
- Circuit breaker
- CQRS
- Event sourcing

---

# 24. Learning Outcome

After completing this project you should understand:

- How microservices communicate
- Why distributed systems are difficult
- Event-driven architecture
- Authentication in distributed systems
- Message brokers
- Docker networking
- Monitoring and tracing
- Service boundaries
- Async vs sync communication

---

# 25. References

## Official Documentation

- Go: https://go.dev/doc/
- Gin: https://gin-gonic.com/
- PostgreSQL: https://www.postgresql.org/docs/
- RabbitMQ: https://www.rabbitmq.com/tutorials.html
- Docker: https://docs.docker.com/
- Prometheus: https://prometheus.io/docs/
- Grafana: https://grafana.com/docs/
- OpenTelemetry: https://opentelemetry.io/docs/
