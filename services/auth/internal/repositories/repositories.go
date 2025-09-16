package repositories

import (
	"context"

	db "github.com/cp-rektmart/aconcert-microservice/auth/db/codegen"
	"github.com/cp-rektmart/aconcert-microservice/auth/internal/entities"
	"github.com/cp-rektmart/aconcert-microservice/auth/internal/jwt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AuthRepository interface {
	GetUser(ctx context.Context, id uuid.UUID) (entities.User, error)
	CreateUser(ctx context.Context, userInput entities.CreateUserInput) (entities.User, error)
	GetUserByProviderEmail(ctx context.Context, provider entities.Provider, email string) (entities.User, error)

	// Redis
	SetUserAuthToken(ctx context.Context, userID uuid.UUID, token entities.CachedTokens) error
	GetUserAuthToken(ctx context.Context, userID uuid.UUID) (entities.CachedTokens, error)
	DeleteUserAuthToken(ctx context.Context, userID uuid.UUID) error
}

type AuthRepositoryImpl struct {
	db          *db.Queries
	redisClient *redis.Client
	jwtConfig   *jwt.Config
}

func NewRepository(db *db.Queries, redisClient *redis.Client, jwtConfig *jwt.Config) AuthRepository {
	return &AuthRepositoryImpl{
		db:          db,
		redisClient: redisClient,
		jwtConfig:   jwtConfig,
	}
}

func (a *AuthRepositoryImpl) CreateUser(ctx context.Context, userInput entities.CreateUserInput) (entities.User, error) {
	panic("unimplemented")
}

func (a *AuthRepositoryImpl) GetUser(ctx context.Context, id uuid.UUID) (entities.User, error) {
	panic("unimplemented")
}

func (a *AuthRepositoryImpl) GetUserByProviderEmail(ctx context.Context, provider entities.Provider, email string) (entities.User, error) {
	panic("unimplemented")
}

func (a *AuthRepositoryImpl) GetUserAuthToken(ctx context.Context, userID uuid.UUID) (entities.CachedTokens, error) {
	panic("unimplemented")
}

func (a *AuthRepositoryImpl) SetUserAuthToken(ctx context.Context, userID uuid.UUID, token entities.CachedTokens) error {
	panic("unimplemented")
}

func (a *AuthRepositoryImpl) DeleteUserAuthToken(ctx context.Context, userID uuid.UUID) error {
	panic("unimplemented")
}
