package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/postgres"
	"github.com/cp-rektmart/aconcert-microservice/pkg/rabbitmq"
	"github.com/cp-rektmart/aconcert-microservice/pkg/redis"
	"github.com/joho/godotenv"
)

const (
	DefaultENVPath = "./.env"
)

type CorsConfig struct {
	AllowedOrigins   string `env:"ALLOWED_ORIGINS"`
	AllowedMethods   string `env:"ALLOWED_METHODS"`
	AllowedHeaders   string `env:"ALLOWED_HEADERS"`
	AllowCredentials bool   `env:"ALLOW_CREDENTIALS"`
}

type AppConfig struct {
	Name         string          `env:"NAME"`
	Port         int             `env:"PORT"`
	Environment  string          `env:"ENVIRONMENT"`
	MaxBodyLimit int             `env:"MAX_BODY_LIMIT"`
	Logger       logger.Config   `envPrefix:"LOGGER_"`
	Postgres     postgres.Config `envPrefix:"POSTGRES_"`
	Cors         CorsConfig      `envPrefix:"CORS_"`
	RabbitMQ     rabbitmq.Config `envPrefix:"RABBITMQ_"`
	Redis        redis.Config    `envPrefix:"REDIS_"`
}

func Load() *AppConfig {
	appConfig := &AppConfig{}
	_ = godotenv.Load()

	if err := env.Parse(appConfig); err != nil {
		panic(err)
	}

	return appConfig
}
