package domain

import (
	"context"

	"github.com/cp-rektmart/aconcert-microservice/auth/internal/entities"
	"github.com/cp-rektmart/aconcert-microservice/auth/internal/repositories"
	"github.com/google/uuid"
)

type AuthDomain interface {
	LoginWithGoogle(ctx context.Context, idToken string) (entities.Token, entities.User, bool, error)
	RefreshToken(ctx context.Context, token string) (entities.Token, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	GetUser(ctx context.Context, userID uuid.UUID) (entities.User, error)
}

type AuthDomainImpl struct {
	repo repositories.AuthRepository
}

func New(repo repositories.AuthRepository) AuthDomain {
	return &AuthDomainImpl{
		repo: repo,
	}
}

func (d *AuthDomainImpl) LoginWithGoogle(ctx context.Context, idToken string) (entities.Token, entities.User, bool, error) {
	panic("unimplemented")
}

func (d *AuthDomainImpl) RefreshToken(ctx context.Context, token string) (entities.Token, error) {
	panic("unimplemented")
}

func (d *AuthDomainImpl) Logout(ctx context.Context, userID uuid.UUID) error {
	panic("unimplemented")
}

func (d *AuthDomainImpl) GetUser(ctx context.Context, userID uuid.UUID) (entities.User, error) {
	panic("unimplemented")
}
