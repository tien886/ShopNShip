package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// DefaultJWTSecret is the dev-only default. All three services MUST use the
// same value (or the same overridden JWT_SECRET env var) so that tokens issued
// by auth-service are accepted by order-service and delivery-service.
const DefaultJWTSecret = "shopnship-dev-secret-change-me"

type Config struct {
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	JWTSecret  string `mapstructure:"JWT_SECRET"`
	ServerPort string `mapstructure:"SERVER_PORT"`
}

func LoadConfig() (*Config, error) {
	// Set defaults
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "auth_db")
	viper.SetDefault("JWT_SECRET", DefaultJWTSecret)
	viper.SetDefault("SERVER_PORT", "8080")
	// Read from .env file
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./auth-service") // Support running from root

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Warn().Err(err).Msg("Error reading config file")
		}
	}

	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	log.Info().Msg("Configuration loaded successfully")
	return &config, nil
}
