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
	_ "github.com/tien886/ShopNShip/delivery-service/docs"
	"github.com/tien886/ShopNShip/delivery-service/internal/config"
	"github.com/tien886/ShopNShip/delivery-service/internal/db"
	"github.com/tien886/ShopNShip/delivery-service/internal/event"
	"github.com/tien886/ShopNShip/delivery-service/internal/handler"
	"github.com/tien886/ShopNShip/delivery-service/internal/middleware"
	"github.com/tien886/ShopNShip/delivery-service/internal/repository"
	"github.com/tien886/ShopNShip/delivery-service/internal/service"
)

// @title           ShopNShip Delivery Service API
// @version         1.0
// @description     Delivery management service for ShopNShip microservices system.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8082
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

	deliveryRepo := repository.NewDeliveryRepository(database)
	deliverySvc := service.NewDeliveryService(deliveryRepo)
	deliveryHandler := handler.NewDeliveryHandler(deliverySvc)

	var consumer *event.EventConsumer
	consumer, err = event.NewEventConsumer(cfg.RabbitMQURL, deliverySvc)
	if err != nil {
		log.Printf("Warning: failed to initialize event consumer: %v", err)
	} else {
		if err := consumer.Start(); err != nil {
			log.Printf("Warning: failed to start consumer: %v", err)
		}
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "up", "service": "delivery-service"})
	})

	// Scalar documentation (public — under /deliveries prefix)
	deliveries := r.Group("/deliveries")
	{
		deliveries.GET("/docs/*any", openapiui.WrapHandler(openapiui.Config{
			SpecURL:      "/deliveries/swagger.json",
			SpecFilePath: "./docs/swagger.json",
			Title:        "ShopNShip Delivery API",
			Theme:        "dark",
		}))
		deliveries.StaticFile("/swagger.json", "./docs/swagger.json")
	}

	protected := r.Group("/deliveries")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret, nil))
	{
		protected.GET("", deliveryHandler.GetDeliveries)
		protected.GET("/:id", deliveryHandler.GetDelivery)
		protected.PATCH("/:id/status", deliveryHandler.UpdateStatus)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Delivery Service starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down delivery-service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown failed: %v", err)
	}

	if consumer != nil {
		if err := consumer.Shutdown(); err != nil {
			log.Printf("failed to shutdown consumer: %v", err)
		}
	}
	log.Println("delivery-service stopped")
}
