package dto

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/auth/internal/entities"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/moonrhythm/validator"
)

type UserResponse struct {
	ID           uuid.UUID         `json:"id" validate:"required"`
	Provider     entities.Provider `json:"provider" validate:"required"`
	Email        string            `json:"email" validate:"required"`
	Firstname    string            `json:"firstname" validate:"required"`
	Lastname     string            `json:"lastname" validate:"required"`
	ProfileImage string            `json:"profileImage" validate:"required"`
	Birthdate    time.Time         `json:"birthdate" validate:"required"`
	Phone        string            `json:"phone" validate:"required"`
	Role         entities.UserRole `json:"role" validate:"required"`
	CreatedAt    time.Time         `json:"createdAt" validate:"required"`
	UpdatedAt    time.Time         `json:"updatedAt" validate:"required"`
	DeletedAt    time.Time         `json:"deletedAt"`
}

type LoginWithProviderRequest struct {
	Provider string `json:"provider"`
	IdToken  string `json:"idToken"`
}

func (l *LoginWithProviderRequest) Parse(c *fiber.Ctx) error {
	if err := c.BodyParser(l); err != nil {
		return errors.Wrap(err, "failed to parse request")
	}

	if err := l.Validate(); err != nil {
		return errors.Wrap(err, "failed to validate request")
	}

	return nil
}

func (l *LoginWithProviderRequest) Validate() error {
	v := validator.New()
	v.Must(l.Provider != "", "provider is required")
	v.Must(l.IdToken != "", "id token is required")

	return errors.WithStack(v.Error())
}

type LoginWithProviderResponse struct {
	AccessToken  string       `json:"accessToken" validate:"required"`
	RefreshToken string       `json:"refreshToken" validate:"required"`
	Exp          int64        `json:"exp" validate:"required"`
	User         UserResponse `json:"user" validate:"required"`
	IsNewUser    bool         `json:"isNewUser" validate:"required"`
}
