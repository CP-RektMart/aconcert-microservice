package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/joho/godotenv"
)

const (
	DefaultENVPath = "./.env"
)

type StripeConfig struct {
	SecretKey     string `env:"SECRET_KEY"`
	ReturnURL     string `env:"RETURN_URL"`
	SigningSecret string `env:"SIGNING_SECRET"`
}

type AppConfig struct {
	Name                     string        `env:"NAME"`
	Port                     int           `env:"PORT"`
	Environment              string        `env:"ENVIRONMENT"`
	MaxBodyLimit             int           `env:"MAX_BODY_LIMIT"`
	Logger                   logger.Config `envPrefix:"LOGGER_"`
	Stripe                   StripeConfig  `envPrefix:"STRIPE_"`
	ReservationClientBaseURL string        `env:"RESERVATION_CLIENT_BASE_URL"`
}

func Load() *AppConfig {
	appConfig := &AppConfig{}
	_ = godotenv.Load()

	if err := env.Parse(appConfig); err != nil {
		panic(err)
	}

	return appConfig
}
