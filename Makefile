# Makefile for ShopNShip Delivery Management Microservice System

COMPOSE=podman compose
AUTH_DIR=auth-service
ORDER_DIR=order-service
DELIVERY_DIR=delivery-service

.PHONY: help up down build rebuild logs ps clean dev infra run-auth run-order run-delivery run-all tidy

.DEFAULT_GOAL := help

help:
	@echo "Available commands:"
	@echo "  make up          - Start all services in background (Podman)"
	@echo "  make down        - Stop all services (Podman)"
	@echo "  make build       - Build all service images"
	@echo "  make rebuild     - Rebuild and restart all services"
	@echo "  make infra       - Start only infrastructure (DBs, RabbitMQ, Redis)"
	@echo "  make dev         - Start infra and run all Go services locally"
	@echo "  make run-auth    - Run Auth Service locally"
	@echo "  make run-order   - Run Order Service locally"
	@echo "  make run-delivery- Run Delivery Service locally"
	@echo "  make tidy        - Run go mod tidy for all services"
	@echo "  make logs        - Follow logs from all services"
	@echo "  make ps          - List running services"
	@echo "  make clean       - Stop services and remove volumes/prune images"

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

build:
	$(COMPOSE) build

rebuild:
	$(COMPOSE) up -d --build

logs:
	$(COMPOSE) logs -f

ps:
	$(COMPOSE) ps

infra:
	$(COMPOSE) up -d postgres-auth postgres-order postgres-delivery rabbitmq redis

dev: infra run-all

run-all:
	@echo "Starting all services locally..."
	make -j 3 run-auth run-order run-delivery

run-auth:
	@echo "Generating Auth Service Swagger docs..."
	cd $(AUTH_DIR) && go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/main.go -o docs
	@echo "Starting Auth Service..."
	cd $(AUTH_DIR) && go run cmd/main.go

run-order:
	@echo "Starting Order Service..."
	cd $(ORDER_DIR) && go run cmd/main.go

run-delivery:
	@echo "Starting Delivery Service..."
	cd $(DELIVERY_DIR) && go run cmd/main.go

tidy:
	@echo "Tidying go.mod for all services..."
	cd $(AUTH_DIR) && go mod tidy
	cd $(ORDER_DIR) && go mod tidy
	cd $(DELIVERY_DIR) && go mod tidy

clean:
	$(COMPOSE) down -v
	@echo "Cleaning up dangling images..."
	podman image prune -f
