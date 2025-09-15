package repositories

import (
	"context"

	"github.com/cp-rektmart/aconcert-microservice/auth/internal/entities"
	"github.com/google/uuid"
)

type AuthRepository interface {
	GetUserAuthToken(ctx context.Context, userID uuid.UUID) (entities.CachedTokens, error)
}
