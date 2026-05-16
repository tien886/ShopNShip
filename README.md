# ShopNShip - Delivery Management Microservice System

A learning-focused microservices project built with **Go**, **PostgreSQL**, **RabbitMQ**, **Redis**, and **NGINX**.

## 🚀 Architecture Overview

- **API Gateway (NGINX)**: Entry point for all requests, routing to internal services.
- **Auth Service**: Manages user registration, login, and JWT-based authentication.
- **Order Service**: Handles order creation and lifecycle management.
- **Delivery Service**: Manages delivery tasks and driver assignments.
- **Infrastructure**: RabbitMQ for async communication and Redis for caching.

## 🛠 Tech Stack

- **Backend**: Go (Gin Framework)
- **Database**: PostgreSQL (Service-per-database pattern)
- **Message Broker**: RabbitMQ
- **Documentation**: Scalar (Accessible at `/docs` for each service)
- **Containerization**: Podman / Docker

## 🚦 Getting Started

### Prerequisites

- **Go 1.22+**
- **Podman** (or Docker)

### Configuration

All three services share a single JWT signing secret — auth-service signs tokens, order-service and delivery-service verify them, so the value **must be identical** everywhere.

Copy `.env.example` to `.env` to override the default secret:

```bash
cp .env.example .env
# then edit JWT_SECRET to a strong random value before deploying
```

If `JWT_SECRET` is not set, all three services fall back to the same dev default (`shopnship-dev-secret-change-me`).

### Running the System

You can use the provided **Makefile** to manage the services:

```bash
# Build and start all services
make up

# Check service status
make ps

# View logs
make logs

# Stop all services
make down
```

### API Documentation

Once the services are running, you can access the **Scalar** API documentation at:

- Auth Service: `http://localhost/auth/docs`
- Order Service: `http://localhost/orders/docs`
- Delivery Service: `http://localhost/deliveries/docs`

## 📁 Project Structure

```text
ShopNShip/
├── auth-service/     # Authentication & Identity
├── order-service/    # Order management
├── delivery-service/ # Delivery logic
├── gateway/          # NGINX configuration
├── shared/           # Common utilities & events
├── docs/             # Project documentation
└── docker-compose.yml
```

## 📜 License

This project is for learning purposes.
