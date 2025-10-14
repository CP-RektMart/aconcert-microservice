package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	Host     string `env:"MONGO_HOST" envDefault:"localhost"`
	Port     int    `env:"MONGO_PORT" envDefault:"27017"`
	User     string `env:"MONGO_USER" envDefault:"root"`
	Password string `env:"MONGO_PASSWORD" envDefault:"password"`
	Database string `env:"MONGO_DB" envDefault:"aconcert"`
}

type Mongo struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func NewMongo(cfg Config) (*Mongo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d", cfg.User, cfg.Password, cfg.Host, cfg.Port)
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(cfg.Database)

	return &Mongo{
		Client: client,
		DB:     db,
	}, nil
}

func (m *Mongo) Disconnect(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}
