package authentication

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/auth/internal/repositories"
	"github.com/cp-rektmart/aconcert-microservice/pkg/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var (
	ErrUnauthorized = errors.New("UNAUTHORIZED")
)

type AuthMiddleware interface {
	Auth(ctx *fiber.Ctx) error
	GetUserIDFromContext(ctx context.Context) (uuid.UUID, error)
}

type authMiddleware struct {
	authRepo repositories.AuthRepository
	config   *jwt.Config
}

func NewAuthMiddleware(authRepo repositories.AuthRepository, config *jwt.Config) AuthMiddleware {
	return &authMiddleware{
		authRepo: authRepo,
		config:   config,
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

func (r *authMiddleware) validateToken(ctx context.Context, bearerToken string) (jwt.JWTentity, error) {
	parsedToken, err := jwt.ParseToken(bearerToken, r.config.AccessTokenSecret)
	if err != nil {
		return jwt.JWTentity{}, errors.Wrap(err, "failed to parse refresh token")
	}

	cachedToken, err := r.authRepo.GetUserAuthToken(ctx, parsedToken.ID)
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
