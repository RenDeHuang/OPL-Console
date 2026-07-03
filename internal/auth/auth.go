package auth

import (
	"errors"
	"time"
)

type Role string
type Status string

var (
	ErrInvalidCredentials = errors.New("invalid_credentials")
	ErrUserDisabled       = errors.New("user_disabled")
	ErrSessionNotFound    = errors.New("session_not_found")
)

const (
	RoleOwner Role = "owner"
	RoleAdmin Role = "admin"

	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
)

type User struct {
	ID     string
	Email  string
	Role   Role
	Status Status
}

type Session struct {
	Token     string
	CSRFToken string
	User      User
	ExpiresAt time.Time
}

func CanAccessOwner(user User) bool {
	return user.Status == StatusActive && (user.Role == RoleOwner || user.Role == RoleAdmin)
}

func CanAccessAdmin(user User) bool {
	return user.Status == StatusActive && user.Role == RoleAdmin
}
