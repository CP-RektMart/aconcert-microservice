package store

import (
	"fmt"
	"log"
	"os"

	db "github.com/cp-rektmart/aconcert-microservice/event/db/codegen"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"golang.org/x/net/context"
)

type Store struct {
	DB    *db.Queries
	Pool  *pgxpool.Pool
	Redis *redis.Client
}

func NewStore(dbConnStr string, redisAddr string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	q := db.New(pool)

	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Store{
		DB:    q,
		Pool:  pool,
		Redis: redisClient,
	}, nil
}

func (s *Store) Close() {
	s.Pool.Close()
	if err := s.Redis.Close(); err != nil {
		log.Printf("failed to close Redis connection: %v", err)
	}
}

// Additional methods for data storage and retrieval can be added here.
