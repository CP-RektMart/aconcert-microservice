package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/rabbitmq"
	"github.com/cp-rektmart/aconcert-microservice/pkg/realtime"
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
	Name        string          `env:"NAME"`
	Environment string          `env:"ENVIRONMENT"`
	Logger      logger.Config   `envPrefix:"LOGGER_"`
	RabbitMQ    rabbitmq.Config `envPrefix:"RABBITMQ_"`
	Realtime    realtime.Config `envPrefix:"REALTIME_"`
}

func Load() *AppConfig {
	appConfig := &AppConfig{}
	_ = godotenv.Load()

	if err := env.Parse(appConfig); err != nil {
		panic(err)
	}

	return appConfig
}
