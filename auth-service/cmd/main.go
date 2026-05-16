package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	openapiui "github.com/PeterTakahashi/gin-openapi/openapiui"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tien886/ShopNShip/auth-service/internal/config"
	"github.com/tien886/ShopNShip/auth-service/internal/db"
	"github.com/tien886/ShopNShip/auth-service/internal/handler"
	"github.com/tien886/ShopNShip/auth-service/internal/middleware"
	"github.com/tien886/ShopNShip/auth-service/internal/repository"
	"github.com/tien886/ShopNShip/auth-service/internal/service"
)

// @title ShopNShip Auth Service API
// @version 1.0
// @description Authentication service for ShopNShip microservices system.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize database
	database, err := db.InitDB(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}

	// Initialize layers
	userRepo := repository.NewUserRepository(database)
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authSvc)

	// Setup router
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "up", "service": "auth-service"})
	})

	// Public routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)

		// Scalar documentation
		auth.GET("/docs/*any", openapiui.WrapHandler(openapiui.Config{
			SpecURL:      "/auth/swagger.json",
			SpecFilePath: "./docs/swagger.json",
			Title:        "ShopNShip Auth API",
			Theme:        "dark",
		}))

		// Serve swagger.json inside the group
		auth.StaticFile("/swagger.json", "./docs/swagger.json")
	}

	// Protected routes
	protected := r.Group("/auth")
	protected.Use(middleware.AuthMiddleware(authSvc))
	{
		protected.GET("/me", authHandler.Me)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Msgf("Auth Service starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	<-quit
	log.Info().Msg("shutting down auth-service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server shutdown failed")
	}
	log.Info().Msg("auth-service stopped")
}
