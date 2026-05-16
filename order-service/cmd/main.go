package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	openapiui "github.com/PeterTakahashi/gin-openapi/openapiui"
	"github.com/gin-gonic/gin"
	_ "github.com/tien886/ShopNShip/order-service/docs"
	"github.com/tien886/ShopNShip/order-service/internal/config"
	"github.com/tien886/ShopNShip/order-service/internal/db"
	"github.com/tien886/ShopNShip/order-service/internal/event"
	"github.com/tien886/ShopNShip/order-service/internal/handler"
	"github.com/tien886/ShopNShip/order-service/internal/middleware"
	"github.com/tien886/ShopNShip/order-service/internal/repository"
	"github.com/tien886/ShopNShip/order-service/internal/service"
)

// @title           ShopNShip Order Service API
// @version         1.0
// @description     Order management service for ShopNShip microservices system.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8081
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.InitDB(cfg)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	producer, err := event.NewEventProducer(cfg.RabbitMQURL)
	if err != nil {
		log.Printf("Warning: failed to initialize event producer: %v", err)
		// We still start the service so HTTP traffic can be served even if
		// RabbitMQ is temporarily unavailable.
	}

	orderRepo := repository.NewOrderRepository(database)
	orderSvc := service.NewOrderService(orderRepo, producer)
	orderHandler := handler.NewOrderHandler(orderSvc)

	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "up", "service": "order-service"})
	})

	// Routes
	orders := r.Group("/orders")
	{
		// Scalar documentation (public)
		orders.GET("/docs/*any", openapiui.WrapHandler(openapiui.Config{
			SpecURL:      "/orders/swagger.json",
			SpecFilePath: "./docs/swagger.json",
			Title:        "ShopNShip Order API",
			Theme:        "dark",
		}))

		// Serve swagger.json inside the group
		orders.StaticFile("/swagger.json", "./docs/swagger.json")
	}

	// Protected routes
	protected := r.Group("/orders")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret, nil))
	{
		protected.POST("", orderHandler.CreateOrder)
		protected.GET("", orderHandler.GetUserOrders)
		protected.GET("/:id", orderHandler.GetOrder)
		protected.PATCH("/:id/status", orderHandler.UpdateStatus)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Order Service starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down order-service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown failed: %v", err)
	}

	if producer != nil {
		if err := producer.Close(); err != nil {
			log.Printf("failed to close producer: %v", err)
		}
	}
	log.Println("order-service stopped")
}
