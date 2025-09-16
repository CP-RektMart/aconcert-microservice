package postgres

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string `env:"HOST"`
	User     string `env:"USER"`
	Password string `env:"PASSWORD"`
	DBName   string `env:"DBNAME"`
	Port     int    `env:"PORT"`
	SSLMode  string `env:"SSLMODE"`
}

func (c Config) String() string {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s", c.Host, c.User, c.Password, c.DBName, c.Port, c.SSLMode)
	return dsn
}

func (c Config) ParseURL() string {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
	return dsn
}

func NewPool(ctx context.Context, conf Config) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(conf.ParseURL())
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse config")
	}
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create connection pool")
	}

	return pool, nil
}
