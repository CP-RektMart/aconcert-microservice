package auth

import (
	"context"

	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/dto"
	"github.com/google/uuid"
)

type AuthService struct {
}

func NewService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (dto.LoginResponse, error) {
	panic("unimplemented")
}

func (s *AuthService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (dto.RefreshTokenResponse, error) {
	panic("unimplemented")
}

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	panic("unimplemented")
}
