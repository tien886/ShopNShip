package config

import (
	"os"

	"github.com/joho/godotenv"
)

// DefaultJWTSecret is the dev-only default. All three services MUST use the
// same value (or the same overridden JWT_SECRET env var) so that tokens issued
// by auth-service are accepted by order-service and delivery-service.
const DefaultJWTSecret = "shopnship-dev-secret-change-me"

type Config struct {
	Port        string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	RabbitMQURL string
	JWTSecret   string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	return &Config{
		Port:        getEnv("PORT", "8082"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "postgres"),
		DBName:      getEnv("DB_NAME", "delivery_db"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://shopnship:shopnship@localhost:5672/"),
		JWTSecret:   getEnv("JWT_SECRET", DefaultJWTSecret),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
