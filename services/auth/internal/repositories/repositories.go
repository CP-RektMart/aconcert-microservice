package repositories

import (
	"context"

	"github.com/cockroachdb/errors"
	db "github.com/cp-rektmart/aconcert-microservice/auth/db/codegen"
	"github.com/cp-rektmart/aconcert-microservice/auth/internal/entities"
	"github.com/cp-rektmart/aconcert-microservice/auth/internal/jwt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

// Postgres

func (r *AuthRepositoryImpl) CreateUser(ctx context.Context, userInput entities.CreateUserInput) (entities.User, error) {
	param := db.CreateUserParams{
		Provider:  userInput.Provider.String(),
		Email:     userInput.Email,
		FirstName: userInput.Firstname,
		LastName:  userInput.Lastname,
		Role:      string(entities.UserRoleUser),
	}

	if userInput.ProfileImage != "" {
		param.ProfileImage = pgtype.Text{
			String: userInput.ProfileImage,
			Valid:  true,
		}
	}

	user, err := r.db.CreateUser(ctx, param)
	if err != nil {
		return entities.User{}, errors.Wrap(err, "can't create user")
	}

	return entities.UserModelToEntity(user), nil
}

func (r *AuthRepositoryImpl) GetUser(ctx context.Context, id uuid.UUID) (entities.User, error) {
	user, err := r.db.GetUser(ctx, ParseUUID(id))
	if err != nil {
		return entities.User{}, errors.Wrap(err, "can't get user")
	}

	return entities.UserModelToEntity(user), nil
}

func (r *AuthRepositoryImpl) GetUserByProviderEmail(ctx context.Context, provider entities.Provider, email string) (entities.User, error) {
	user, err := r.db.GetUserByProviderEmail(ctx, db.GetUserByProviderEmailParams{
		Provider: provider.String(),
		Email:    email,
	})
	if err != nil {
		return entities.User{}, errors.Wrap(err, "can't get user by provider and email")
	}

	return entities.UserModelToEntity(user), nil
}

// Redis

func (r *AuthRepositoryImpl) GetUserAuthToken(ctx context.Context, userID uuid.UUID) (entities.CachedTokens, error) {
	panic("unimplemented")
}

func (r *AuthRepositoryImpl) SetUserAuthToken(ctx context.Context, userID uuid.UUID, token entities.CachedTokens) error {
	panic("unimplemented")
}

func (r *AuthRepositoryImpl) DeleteUserAuthToken(ctx context.Context, userID uuid.UUID) error {
	panic("unimplemented")
}
