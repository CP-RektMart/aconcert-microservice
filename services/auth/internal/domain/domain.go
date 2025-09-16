package domain

import (
	"context"

	"github.com/cp-rektmart/aconcert-microservice/auth/internal/entities"
	"github.com/google/uuid"
)

type Domain interface {
	LoginWithGoogle(ctx context.Context, idToken string) (entities.Token, entities.User, bool, error)
	RefreshToken(ctx context.Context, token string) (entities.Token, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	GetUser(ctx context.Context, userID uuid.UUID) (entities.User, error)
}
