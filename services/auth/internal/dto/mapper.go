package dto

import (
	"time"

	"github.com/cp-rektmart/aconcert-microservice/auth/internal/entities"
)

func UserEntityToDTO(user entities.User) UserResponse {
	result := UserResponse{
		ID:        user.ID,
		Provider:  user.Provider,
		Email:     user.Email,
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	if user.ProfileImage != "" {
		result.ProfileImage = &user.ProfileImage
	}

	if user.Birthdate != (time.Time{}) {
		result.Birthdate = &user.Birthdate
	}

	if user.Phone != "" {
		result.Phone = &user.Phone
	}

	if user.DeletedAt != nil {
		result.DeletedAt = user.DeletedAt
	}

	return result
}
