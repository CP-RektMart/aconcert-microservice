package dto

import (
	"time"

	"github.com/google/uuid"
)

type LoginRequest struct {
	Provider string `json:"provider"`
	IdToken  string `json:"idToken"`
}

type User struct {
	ID       string `json:"id" validate:"required"`
	Provider string `json:"provider" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Fullname string `json:"fullname" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
	Role     string `json:"role" validate:"required"`
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

type UserResponse struct {
	ID           uuid.UUID  `json:"id" validate:"required"`
	Provider     string     `json:"provider" validate:"required"`
	Email        string     `json:"email" validate:"required"`
	Firstname    string     `json:"firstname" validate:"required"`
	Lastname     string     `json:"lastname" validate:"required"`
	ProfileImage *string    `json:"profileImage" validate:"required"`
	Birthdate    *time.Time `json:"birthdate" validate:"required"`
	Phone        *string    `json:"phone" validate:"required"`
	Role         string     `json:"role" validate:"required"`
	CreatedAt    time.Time  `json:"createdAt" validate:"required"`
	UpdatedAt    time.Time  `json:"updatedAt" validate:"required"`
	DeletedAt    *time.Time `json:"deletedAt"`
}
