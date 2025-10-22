package authentication

import (
	"context"
	"encoding/json"

	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/pkg/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const AuthTokenKey = "auth:token"

var ErrUnauthorized = errors.New("UNAUTHORIZED")

type AuthMiddleware interface {
	Auth(ctx *fiber.Ctx) error
	GetUserIDFromContext(ctx context.Context) (uuid.UUID, error)
}

type authMiddleware struct {
	config          *jwt.Config
	authRedisClient *redis.Client
}

func NewAuthMiddleware(config *jwt.Config, authRedisClient *redis.Client) AuthMiddleware {
	return &authMiddleware{
		config:          config,
		authRedisClient: authRedisClient,
	}
}

func (r *authMiddleware) Auth(ctx *fiber.Ctx) error {
	tokenByte := ctx.GetReqHeaders()["Authorization"]
	if len(tokenByte) == 0 {
		return ErrUnauthorized
	}

	if len(tokenByte[0]) < 7 {
		return ErrUnauthorized
	}

	bearerToken := tokenByte[0][7:]
	if len(bearerToken) == 0 {
		return ErrUnauthorized
	}

	claims, err := r.validateToken(ctx.UserContext(), bearerToken)
	if err != nil {
		return ErrUnauthorized
	}

	userContext := r.withUserID(ctx.UserContext(), claims.ID)
	ctx.SetUserContext(userContext)

	return ctx.Next()
}

type tokenUID struct {
	AccessUID  uuid.UUID `msgpack:"access_uid"`
	RefreshUID uuid.UUID `msgpack:"refresh_uid"`
}

func (r *authMiddleware) getTokenKey(userID uuid.UUID) string {
	return AuthTokenKey + ":" + userID.String()
}

func (r *authMiddleware) GetUserAuthToken(ctx context.Context, userID uuid.UUID) (jwt.CachedTokens, error) {
	redisToken, err := r.authRedisClient.Get(ctx, r.getTokenKey(userID)).Bytes()
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

func (r *authMiddleware) validateToken(ctx context.Context, bearerToken string) (jwt.JWTentity, error) {
	parsedToken, err := jwt.ParseToken(bearerToken, r.config.AccessTokenSecret)
	if err != nil {
		return jwt.JWTentity{}, errors.Wrap(err, "failed to parse refresh token")
	}

	cachedToken, err := r.GetUserAuthToken(ctx, parsedToken.ID)
	if err != nil {
		return jwt.JWTentity{}, errors.Wrap(err, "failed to get cached token")
	}

	err = jwt.ValidateToken(cachedToken, parsedToken, false)
	if err != nil {
		return jwt.JWTentity{}, errors.Wrap(err, "failed to validate refresh token")
	}

	return parsedToken, nil

}

type userIDContext struct{}

func (r *authMiddleware) withUserID(ctx context.Context,
	userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDContext{}, userID)
}

func (r *authMiddleware) GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(userIDContext{}).(uuid.UUID)

	if !ok {
		return uuid.UUID{}, errors.New("failed to get user id from context")
	}

	return userID, nil
}
