package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/joho/godotenv"
)

type CorsConfig struct {
	AllowedOrigins   string `env:"ALLOWED_ORIGINS"`
	AllowedMethods   string `env:"ALLOWED_METHODS"`
	AllowedHeaders   string `env:"ALLOWED_HEADERS"`
	AllowCredentials bool   `env:"ALLOW_CREDENTIALS"`
}

type AppConfig struct {
	Name         string        `env:"NAME"`
	Port         int           `env:"PORT"`
	Environment  string        `env:"ENVIRONMENT"`
	MaxBodyLimit int           `env:"MAX_BODY_LIMIT"`
	Cors         CorsConfig    `envPrefix:"CORS_"`
	Logger       logger.Config `envPrefix:"LOGGER_"`
}

func Load() *AppConfig {
	appConfig := &AppConfig{}
	_ = godotenv.Load()

	if err := env.Parse(appConfig); err != nil {
		panic(err)
	}

	return appConfig
}
