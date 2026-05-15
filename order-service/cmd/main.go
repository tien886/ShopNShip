package main

import (
	"log"

	openapiui "github.com/PeterTakahashi/gin-openapi/openapiui"
	"github.com/gin-gonic/gin"
	"github.com/tien886/ShopNShip/order-service/internal/config"
	"github.com/tien886/ShopNShip/order-service/internal/db"
	_ "github.com/tien886/ShopNShip/order-service/docs"
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
		// We might still want to run the service even if RabbitMQ is down,
		// depending on requirements.
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

	log.Printf("Order Service starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
