package config

type AppConfig struct {
	Name         string `env:"NAME"`
	Port         int    `env:"PORT"`
	Environment  string `env:"ENVIRONMENT"`
	MaxBodyLimit int    `env:"MAX_BODY_LIMIT"`
}
