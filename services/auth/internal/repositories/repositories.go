package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
	db "github.com/cp-rektmart/aconcert-microservice/auth/db/codegen"
	"github.com/cp-rektmart/aconcert-microservice/auth/internal/entities"
	"github.com/cp-rektmart/aconcert-microservice/auth/internal/errs"
	"github.com/cp-rektmart/aconcert-microservice/pkg/jwt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

const (
	AuthTokenKey = "auth:token"
)

type AuthRepository interface {
	GetUser(ctx context.Context, id uuid.UUID) (entities.User, error)
	CreateUser(ctx context.Context, userInput entities.CreateUserInput) (entities.User, error)
	GetUserByProviderEmail(ctx context.Context, provider entities.Provider, email string) (entities.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, updateInput entities.UpdateUserInput) (entities.User, error)

	// Redis
	SetUserAuthToken(ctx context.Context, userID uuid.UUID, token jwt.CachedTokens) error
	GetUserAuthToken(ctx context.Context, userID uuid.UUID) (jwt.CachedTokens, error)
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
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.User{}, errs.ErrNotFound
		}
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
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.User{}, errs.ErrNotFound
		}
		return entities.User{}, errors.Wrap(err, "can't get user by provider and email")
	}

	return entities.UserModelToEntity(user), nil
}

// Redis
type tokenUID struct {
	AccessUID  uuid.UUID `msgpack:"access_uid"`
	RefreshUID uuid.UUID `msgpack:"refresh_uid"`
}

func (r *AuthRepositoryImpl) getTokenKey(userID uuid.UUID) string {
	return AuthTokenKey + ":" + userID.String()
}

func (r *AuthRepositoryImpl) GetUserAuthToken(ctx context.Context, userID uuid.UUID) (jwt.CachedTokens, error) {
	redisToken, err := r.redisClient.Get(ctx, r.getTokenKey(userID)).Bytes()
	if err != nil {
		return jwt.CachedTokens{}, errors.Wrap(err, "can't get token")
	}

	cachedToken := tokenUID{}
	if err = json.Unmarshal(redisToken, &cachedToken); err != nil {
		return jwt.CachedTokens{}, errors.Wrap(err, "can't unmarshal cached token")
	}

	return jwt.CachedTokens{
		AccessUID:  cachedToken.AccessUID,
		RefreshUID: cachedToken.RefreshUID,
	}, nil
}

func (r *AuthRepositoryImpl) SetUserAuthToken(ctx context.Context, userID uuid.UUID, token jwt.CachedTokens) error {
	cachedToken, err := json.Marshal(tokenUID{
		AccessUID:  token.AccessUID,
		RefreshUID: token.RefreshUID,
	})
	if err != nil {
		return errors.Wrap(err, "can't marshal cached token")
	}

	err = r.redisClient.Set(ctx, r.getTokenKey(userID), string(cachedToken), time.Second*time.Duration(r.jwtConfig.AutoLogout)).Err()
	if err != nil {
		return errors.Wrap(err, "can't set token")
	}

	return nil
}

func (r *AuthRepositoryImpl) DeleteUserAuthToken(ctx context.Context, userID uuid.UUID) error {
	err := r.redisClient.Del(ctx, r.getTokenKey(userID)).Err()
	if err != nil {
		return errors.Wrap(err, "can't delete token")
	}

	return nil
}

func (r *AuthRepositoryImpl) UpdateUser(ctx context.Context, userID uuid.UUID, updateInput entities.UpdateUserInput) (entities.User, error) {
	param := db.UpdateUserParams{
		ID:        ParseUUID(userID),
		FirstName: updateInput.Firstname,
		LastName:  updateInput.Lastname,
	}

	if updateInput.ProfileImage != "" {
		param.ProfileImage = pgtype.Text{
			String: updateInput.ProfileImage,
			Valid:  true,
		}
	}

	if updateInput.Birthdate.IsZero() == false {
		param.BirthDate = pgtype.Date{
			Time:  updateInput.Birthdate,
			Valid: true,
		}
	}

	if updateInput.Phone != "" {
		param.Phone = pgtype.Text{
			String: updateInput.Phone,
			Valid:  true,
		}
	}

	user, err := r.db.UpdateUser(ctx, param)
	if err != nil {
		return entities.User{}, errors.Wrap(err, "can't update user")
	}

	return entities.UserModelToEntity(user), nil
}
