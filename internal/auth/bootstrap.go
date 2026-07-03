package auth

import (
	"encoding/json"
	"errors"
	"strings"
)

const (
	DefaultAdminEmail    = "admin@opl.local"
	DefaultAdminPassword = "OplAdminPass2026!"
)

type BootstrapUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     Role   `json:"role"`
	Name     string `json:"name"`
}

func BootstrapUsersFromJSON(raw string) ([]BootstrapUser, error) {
	if strings.TrimSpace(raw) == "" {
		return []BootstrapUser{{
			ID:       "usr-admin-bootstrap",
			Email:    DefaultAdminEmail,
			Password: DefaultAdminPassword,
			Role:     RoleAdmin,
			Name:     "OPL Admin",
		}}, nil
	}

	var users []BootstrapUser
	if err := json.Unmarshal([]byte(raw), &users); err != nil {
		return nil, err
	}
	for _, user := range users {
		if strings.TrimSpace(user.ID) == "" ||
			strings.TrimSpace(user.Email) == "" ||
			strings.TrimSpace(user.Password) == "" ||
			(user.Role != RoleOwner && user.Role != RoleAdmin) {
			return nil, errors.New("invalid_bootstrap_user")
		}
	}
	return users, nil
}
