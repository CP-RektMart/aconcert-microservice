package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/cp-rektmart/aconcert-microservice/pkg/jwt"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/redis"
	"github.com/joho/godotenv"
)

type CorsConfig struct {
	AllowedOrigins   string `env:"ALLOWED_ORIGINS"`
	AllowedMethods   string `env:"ALLOWED_METHODS"`
	AllowedHeaders   string `env:"ALLOWED_HEADERS"`
	AllowCredentials bool   `env:"ALLOW_CREDENTIALS"`
}

type AppConfig struct {
	Name                  string        `env:"NAME"`
	Port                  int           `env:"PORT"`
	Environment           string        `env:"ENVIRONMENT"`
	MaxBodyLimit          int           `env:"MAX_BODY_LIMIT"`
	Cors                  CorsConfig    `envPrefix:"CORS_"`
	Logger                logger.Config `envPrefix:"LOGGER_"`
	JWT                   jwt.Config    `envPrefix:"JWT_"`
	AuthRedis             redis.Config  `envPrefix:"AUTHREDIS_"`
	AuthClientBaseURL     string        `env:"AUTH_CLIENT_BASE_URL"`
	EventClientBaseURL    string        `env:"EVENT_CLIENT_BASE_URL"`
	LocationClientBaseURL string        `env:"LOCATION_CLIENT_BASE_URL"`
}

func Load() *AppConfig {
	appConfig := &AppConfig{}
	_ = godotenv.Load()

	if err := env.Parse(appConfig); err != nil {
		panic(err)
	}

	return appConfig
}
