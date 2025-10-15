package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/cp-rektmart/aconcert-microservice/pkg/awss3"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/postgres"
	"github.com/cp-rektmart/aconcert-microservice/pkg/rabbitmq"
	"github.com/cp-rektmart/aconcert-microservice/pkg/redis"
	"github.com/joho/godotenv"
)

const (
	DefaultENVPath = "./.env"
)

type AppConfig struct {
	Name        string          `env:"NAME"`
	Port        int             `env:"PORT"`
	Environment string          `env:"ENVIRONMENT"`
	Logger      logger.Config   `envPrefix:"LOGGER_"`
	Postgres    postgres.Config `envPrefix:"POSTGRES_"`
	Redis       redis.Config    `envPrefix:"REDIS_"`
	S3          awss3.Config    `envPrefix:"S3_"`
	RabbitMQ    rabbitmq.Config `envPrefix:"RABBITMQ_"`
}

func Load() *AppConfig {
	appConfig := &AppConfig{}
	_ = godotenv.Load()

	if err := env.Parse(appConfig); err != nil {
		panic(err)
	}

	return appConfig
}
