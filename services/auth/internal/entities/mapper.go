package entities

import (
	db "github.com/cp-rektmart/aconcert-microservice/auth/db/codegen"
	"github.com/google/uuid"
)

func UserModelToEntity(user db.User) User {
	result := User{
		ID:        uuid.UUID(user.ID.Bytes),
		Provider:  Provider(user.Provider),
		Email:     user.Email,
		Firstname: user.FirstName,
		Lastname:  user.LastName,
		Role:      UserRole(user.Role),
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}

	if user.ProfileImage.Valid {
		result.ProfileImage = user.ProfileImage.String
	}

	if user.Phone.Valid {
		result.Phone = user.Phone.String
	}

	if user.BirthDate.Valid {
		result.Birthdate = user.BirthDate.Time
	}

	return result
}
