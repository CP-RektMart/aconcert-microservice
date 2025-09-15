package entities

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleUnknown UserRole = ""
	UserRoleAdmin   UserRole = "ADMIN"
	UserRoleStaff   UserRole = "STAFF"
	UserRoleUser    UserRole = "USER"
)

type Provider string

const (
	ProviderUnknown Provider = ""
	ProviderGoogle  Provider = "GOOGLE"
)

func (p Provider) String() string {
	return string(p)
}

type User struct {
	ID        uuid.UUID
	Provider  Provider
	Email     string
	Fullname  string
	Phone     string
	Role      UserRole
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u User) String() string {
	return u.Fullname
}

type CachedTokens struct {
	AccessUID  uuid.UUID
	RefreshUID uuid.UUID
}
