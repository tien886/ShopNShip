# ShopNShip — Delivery Management Microservices

A learning-focused **Go microservices** project with **PostgreSQL**, **RabbitMQ**, **Redis**, **NGINX**, and **JWT authentication**.

## Architecture

```
                    Client
                      │
                   NGINX (:80)
                      │
       ┌──────────────┼──────────────┐
       ▼              ▼              ▼
  auth-service   order-service  delivery-service
    :8080           :8081           :8082
       │              │               │
   postgres-     postgres-       postgres-
   auth (5433)   order (5434)    delivery (5435)
       │              │
       └──────┬───────┘
              │
          RabbitMQ (5672 / 15672)
```

**Services**

| Service | Port | Role |
|---|---|---|
| `auth-service` | `:8080` | Register, login, refresh, JWT issuance (HS256) |
| `order-service` | `:8081` | Create orders, list orders, update order status |
| `delivery-service` | `:8082` | Receive orders via RabbitMQ, manage delivery lifecycle |
| `gateway` (NGINX) | `:80` | Route `/auth/*`, `/orders/*`, `/deliveries/*` to upstream services |
| `rabbitmq` | `:5672` / `:15672` | Topic exchange `order.events` — `order.created` routing key |
| `postgres-*` | `5433`, `5434`, `5435` | One database per service |
| `redis` | `:6379` | Cache layer (reserved for future use) |

**Business flow**

1. `POST /auth/register` → `POST /auth/login` → receive JWT
2. `POST /orders` (Bearer JWT) → order saved → `OrderCreated` event published
3. `delivery-service` consumes event → creates a `PENDING` delivery
4. Update order/delivery status via `PATCH /orders/:id/status` and `PATCH /deliveries/:id/status`

---

## Prerequisites

- **Go 1.22+** (only needed for local dev mode)
- **Docker** or **Podman** + `docker compose` / `podman compose`
- `make` (optional — all commands below have `docker compose` equivalents)

---

## 1. Download

```bash
git clone https://github.com/tien886/ShopNShip.git
cd ShopNShip
```

---

## 2. Configuration

All three services share one JWT signing secret. Auth-service signs tokens; order-service and delivery-service verify them. The value **must be identical** everywhere.

Copy the example file and optionally change the secret:

```bash
cp .env.example .env
```

`.env`
```ini
JWT_SECRET=shopnship-dev-secret-change-me
```

> **Note:** `docker-compose` automatically reads `.env` in the project root. `.env.example` itself is **not** loaded — you must copy it.
>
> **Production:** Change `JWT_SECRET` to a strong random string. If `.env` is missing, all services fall back to the same dev default.

---

## 3. Run with Docker Compose (recommended)

This builds images, starts infrastructure, and waits for healthchecks before bringing up app services.

```bash
# Start everything
docker compose up -d --build

# Or with Podman
podman compose up -d --build
```

Wait ~30s for RabbitMQ and Postgres to become healthy, then verify:

```bash
# Health checks
curl http://localhost/auth/health
curl http://localhost/orders/health
curl http://localhost/deliveries/health

# List running containers
docker compose ps
```

### Makefile shortcuts

```bash
make up      # docker compose up -d
make down    # docker compose down
make build   # docker compose build
make rebuild # docker compose up -d --build
make logs    # docker compose logs -f
make ps      # docker compose ps
make clean   # docker compose down -v + image prune
```

---

## 4. Run locally (dev mode — no containers for Go services)

Use this when you want to iterate on Go code without rebuilding images.

**Step 1 — Start infrastructure only**

```bash
docker compose up -d postgres-auth postgres-order postgres-delivery rabbitmq redis
```

**Step 2 — Generate auth-service Swagger docs**

```bash
cd auth-service
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/main.go -o docs
```

**Step 3 — Run each service in separate terminals**

```bash
# Terminal 1
cd auth-service && go run cmd/main.go

# Terminal 2
cd order-service && go run cmd/main.go

# Terminal 3
cd delivery-service && go run cmd/main.go
```

Or use the Makefile:

```bash
make infra   # starts DBs + RabbitMQ + Redis
make dev     # starts infra + runs all three services locally
make tidy    # run `go mod tidy` in all three services
```

> **Note:** In dev mode you hit the services directly on their ports (`:8080`, `:8081`, `:8082`). The NGINX gateway (`:80`) is not needed.
>
> **Warning:** `make dev` runs `go run` for all three services in parallel with `make -j`. Output will be interleaved in one terminal. For cleaner logs, run each service in a separate terminal (Step 3 above).

---

## 5. API Quick Start

### 5.1 Register a user

```bash
curl -X POST http://localhost/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"secret123","full_name":"Alice"}'
```

### 5.2 Login

```bash
curl -X POST http://localhost/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"secret123"}'
```

Response:
```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ..."
}
```

### 5.3 Create an order

```bash
export TOKEN="<access_token>"

curl -X POST http://localhost/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      { "item_name": "Widget", "quantity": 2, "price": 29.99 },
      { "item_name": "Gadget", "quantity": 1, "price": 49.50 }
    ]
  }'
```

Response:
```json
{
  "id": "438f2bd9-dbe8-4f6a-9120-005a34b23a67",
  "user_id": 1,
  "status": "PENDING",
  "total_price": 109.48,
  "items": [
    {
      "id": "52903f1a-80ec-42f2-b621-5fc5b201235d",
      "order_id": "438f2bd9-dbe8-4f6a-9120-005a34b23a67",
      "item_name": "Widget",
      "quantity": 2,
      "price": 29.99,
      "created_at": "2026-05-16T09:41:38.321124+07:00"
    },
    {
      "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c9",
      "order_id": "438f2bd9-dbe8-4f6a-9120-005a34b23a67",
      "item_name": "Gadget",
      "quantity": 1,
      "price": 49.50,
      "created_at": "2026-05-16T09:41:38.321124+07:00"
    }
  ],
  "created_at": "2026-05-16T09:41:38.317054+07:00",
  "updated_at": "2026-05-16T09:41:38.317054+07:00"
}
```

> The handler currently returns the full `model.Order` (including `items[].id`, `items[].order_id`, `items[].created_at`, and `updated_at`).

### 5.4 List your orders

```bash
curl http://localhost/orders \
  -H "Authorization: Bearer $TOKEN"
```

### 5.5 Update order status

```bash
curl -X PATCH http://localhost/orders/438f2bd9-dbe8-4f6a-9120-005a34b23a67/status \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"PAID"}'
```

### 5.6 List deliveries (auto-created after order)

```bash
curl http://localhost/deliveries \
  -H "Authorization: Bearer $TOKEN"
```

Response:
```json
[
  {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "order_id": "438f2bd9-dbe8-4f6a-9120-005a34b23a67",
    "user_id": 1,
    "status": "PENDING",
    "created_at": "...",
    "updated_at": "..."
  }
]
```

### 5.7 Update delivery status

```bash
curl -X PATCH http://localhost/deliveries/a1b2c3d4-e5f6-7890-abcd-ef1234567890/status \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"ASSIGNED"}'
```

**Delivery status transitions**

| From | Allowed To |
|---|---|
| `PENDING` | `ASSIGNED`, `CANCELLED` |
| `ASSIGNED` | `IN_TRANSIT`, `CANCELLED` |
| `IN_TRANSIT` | `DELIVERED`, `CANCELLED` |

### 5.8 Refresh token

```bash
curl -X POST http://localhost/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

### 5.9 Study event flow logs (Order -> RabbitMQ -> Delivery)

Run this in another terminal while creating a new order:

```bash
docker compose logs -f order-service delivery-service
```

Look for log lines containing `[FLOW][ORDER->DELIVERY]` and follow these steps:

1. `STEP 1/7 [ORDER-SVC]` request enters order service.
2. `STEP 2/7 [ORDER-SVC]` order is stored in order DB.
3. `STEP 3/7 [ORDER-SVC]/[PRODUCER]` event payload is prepared.
4. `STEP 4/7 [ORDER-SVC]/[PRODUCER]` event is published to `order.events` with `order.created`.
5. `STEP 5/7 [CONSUMER]` delivery service receives RabbitMQ message.
6. `STEP 6/7 [CONSUMER]/[DELIVERY-SVC]` message is handed to delivery business logic.
7. `STEP 7/7 [DELIVERY-SVC]/[CONSUMER]` delivery row is saved and message is ACKed.

---

## 6. API Documentation (Scalar)

> **Note:** Docs routes are mounted inside each service's path prefix (e.g. `/auth/docs/` in auth-service). Because the NGINX gateway strips that prefix when proxying, the docs URLs **do not work through the gateway** in the current configuration. Access them directly via the service ports below.

| Service | URL |
|---|---|
| Auth | `http://localhost:8080/auth/docs/` |
| Order | `http://localhost:8081/orders/docs/` |
| Delivery | `http://localhost:8082/deliveries/docs/` |

> **Trailing slash required.** `http://localhost:8080/auth/docs/` works; `.../auth/docs` (no slash) may 404 or redirect depending on Gin version.

---

## 7. Project Structure

```text
ShopNShip/
├── auth-service/
│   ├── cmd/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── db/
│   │   ├── dto/
│   │   ├── handler/
│   │   ├── middleware/
│   │   ├── model/
│   │   ├── repository/
│   │   └── service/
│   ├── docs/
│   ├── Dockerfile
│   └── go.mod
├── order-service/
│   ├── cmd/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── db/
│   │   ├── dto/
│   │   ├── event/
│   │   ├── handler/
│   │   ├── middleware/
│   │   ├── model/
│   │   ├── repository/
│   │   └── service/
│   ├── docs/
│   ├── Dockerfile
│   └── go.mod
├── delivery-service/
│   ├── cmd/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── db/
│   │   ├── dto/
│   │   ├── event/
│   │   ├── handler/
│   │   ├── middleware/
│   │   ├── model/
│   │   ├── repository/
│   │   └── service/
│   ├── docs/
│   ├── Dockerfile
│   └── go.mod
├── gateway/
│   └── nginx.conf
├── .env.example
├── docker-compose.yml
└── Makefile
```

---

## 8. Troubleshooting

### `authorization header is required` on docs

The docs endpoints (`/auth/docs/`, `/orders/docs/`, `/deliveries/docs/`) are **public** in the current code. If you get an auth error, the running binary is stale. Rebuild:

```bash
docker compose up -d --build delivery-service
```

### `failed to connect to RabbitMQ`

Services retry RabbitMQ for ~30 seconds at startup. If you see warnings in logs, wait a bit and check:

```bash
docker compose logs rabbitmq
```

### Postgres connection refused

Make sure `DB_PORT` inside containers is `5432` (the Postgres container port), not the host-mapped port (`5433`, `5434`, `5435`). This is already correct in the current configs.

### Order creation succeeds but no delivery appears

Check the delivery-service logs:

```bash
docker compose logs -f delivery-service
```

The consumer listens on queue `delivery.order.created` bound to exchange `order.events` with routing key `order.created`.

---

## 9. Environment Variables

| Variable | Default | Used By |
|---|---|---|
| `JWT_SECRET` | `shopnship-dev-secret-change-me` | auth, order, delivery |
| `DB_HOST` | `postgres-{service}` | each service |
| `DB_PORT` | `5432` | each service |
| `DB_USER` | `postgres` | each service |
| `DB_PASSWORD` | `postgres` | each service |
| `DB_NAME` | `{service}_db` | each service |
| `RABBITMQ_URL` | `amqp://shopnship:shopnship@rabbitmq:5672/` | order, delivery |

---

## 10. License

This project is for learning and educational purposes.
