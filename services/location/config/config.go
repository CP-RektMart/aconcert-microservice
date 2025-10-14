package config

import (
	"github.com/caarlos0/env/v9"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/mongodb"
	"github.com/joho/godotenv"
)

const (
	DefaultENVPath = "./.env"
)


type AppConfig struct {
	Name        string        `env:"NAME" envDefault:"aconcert-location-service"`
	Port        int           `env:"PORT" envDefault:"8082"`
	Environment string        `env:"ENVIRONMENT" envDefault:"development"`
	Logger      logger.Config `envPrefix:"LOGGER_"`
	Mongo       mongodb.Config
}

func Load() *AppConfig {
	appConfig := &AppConfig{}
	_ = godotenv.Load(DefaultENVPath)

	if err := env.Parse(appConfig); err != nil {
		panic(err)
	}

	return appConfig
}
