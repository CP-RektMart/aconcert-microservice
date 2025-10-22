package dto

import "github.com/google/uuid"

type LoginRequest struct {
	Provider string `json:"provider"`
	IdToken  string `json:"idToken"`
}

type LoginResponse struct {
	AccessToken  string `json:"accessToken" validate:"required"`
	RefreshToken string `json:"refreshToken" validate:"required"`
	Exp          int64  `json:"exp" validate:"required"`
	User         User   `json:"user" validate:"required"`
	IsNewUser    bool   `json:"isNewUser" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"accessToken" validate:"required"`
	RefreshToken string `json:"refreshToken" validate:"required"`
	Exp          int64  `json:"exp" validate:"required"`
}

type LogoutRequest struct {
	UserID uuid.UUID `json:"userId" validate:"required"`
}

type GetProfileRequest struct {
	UserID uuid.UUID `json:"userId" validate:"required"`
}
