package domain

import (
	"context"

	"github.com/cp-rektmart/aconcert-microservice/auth/internal/entities"
)

type Domain interface {
	LoginWithGoogle(ctx context.Context, idToken string) (entities.Token, entities.User, bool, error)
}
