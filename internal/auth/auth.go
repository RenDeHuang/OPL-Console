package auth

type Role string
type Status string

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

func CanAccessOwner(user User) bool {
	return user.Status == StatusActive && (user.Role == RoleOwner || user.Role == RoleAdmin)
}

func CanAccessAdmin(user User) bool {
	return user.Status == StatusActive && user.Role == RoleAdmin
}
