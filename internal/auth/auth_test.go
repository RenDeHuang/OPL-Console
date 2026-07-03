package auth

import "testing"

func TestCanAccessAdmin(t *testing.T) {
	admin := User{ID: "usr-admin", Role: RoleAdmin, Status: StatusActive}
	owner := User{ID: "usr-owner", Role: RoleOwner, Status: StatusActive}

	if !CanAccessAdmin(admin) {
		t.Fatal("admin should access admin routes")
	}
	if CanAccessAdmin(owner) {
		t.Fatal("owner should not access admin routes")
	}
}

func TestDisabledUserCannotAccessOwner(t *testing.T) {
	user := User{ID: "usr-disabled", Role: RoleOwner, Status: StatusDisabled}
	if CanAccessOwner(user) {
		t.Fatal("disabled user should not access owner routes")
	}
}
