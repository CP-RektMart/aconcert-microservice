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

type StripeConfig struct {
	SecretKey string `env:"SECRET_KEY"`
	ReturnURL string `env:"RETURN_URL"`
}

type AppConfig struct {
	Name               string          `env:"NAME"`
	Port               int             `env:"PORT"`
	Environment        string          `env:"ENVIRONMENT"`
	Logger             logger.Config   `envPrefix:"LOGGER_"`
	Postgres           postgres.Config `envPrefix:"POSTGRES_"`
	Redis              redis.Config    `envPrefix:"REDIS_"`
	Stripe             StripeConfig    `envPrefix:"STRIPE_"`
	RabbitMQ           rabbitmq.Config `envPrefix:"RABBITMQ_"`
	EventClientBaseURL string          `env:"EVENT_CLIENT_BASE_URL"`
}

func Load() *AppConfig {
	appConfig := &AppConfig{}
	_ = godotenv.Load()

	if err := env.Parse(appConfig); err != nil {
		panic(err)
	}

	return appConfig
}
