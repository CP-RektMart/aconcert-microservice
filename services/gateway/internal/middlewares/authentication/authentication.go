package authentication

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("INVALID_TOKEN")
)

type AuthMiddleware interface {
	Auth(ctx *fiber.Ctx) error
	AuthAdmin(ctx *fiber.Ctx) error
	GetUserIDFromContext(ctx context.Context) (uuid.UUID, error)
}
