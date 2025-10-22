package dto

type User struct {
	ID       string `json:"id" validate:"required"`
	Provider string `json:"provider" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Fullname string `json:"fullname" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
	Role     string `json:"role" validate:"required"`
}
