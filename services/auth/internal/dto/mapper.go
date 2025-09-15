package dto

import "github.com/cp-rektmart/aconcert-microservice/auth/internal/entities"

func UserEntityToDTO(user entities.User) UserResponse {
	result := UserResponse{
		ID:           user.ID,
		Provider:     user.Provider,
		Email:        user.Email,
		Firstname:    user.Firstname,
		Lastname:     user.Lastname,
		ProfileImage: user.ProfileImage,
		Birthdate:    user.Birthdate,
		Phone:        user.Phone,
		Role:         user.Role,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	if user.DeletedAt != nil {
		result.DeletedAt = *user.DeletedAt
	}

	return result
}
