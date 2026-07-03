package auth

import "testing"

func TestCanAccessOwner(t *testing.T) {
	tests := []struct {
		name string
		user User
		want bool
	}{
		{
			name: "active owner can access owner routes",
			user: User{ID: "usr-owner", Role: RoleOwner, Status: StatusActive},
			want: true,
		},
		{
			name: "active admin can access owner routes",
			user: User{ID: "usr-admin", Role: RoleAdmin, Status: StatusActive},
			want: true,
		},
		{
			name: "disabled owner cannot access owner routes",
			user: User{ID: "usr-disabled-owner", Role: RoleOwner, Status: StatusDisabled},
			want: false,
		},
		{
			name: "disabled admin cannot access owner routes",
			user: User{ID: "usr-disabled-admin", Role: RoleAdmin, Status: StatusDisabled},
			want: false,
		},
		{
			name: "unknown role cannot access owner routes",
			user: User{ID: "usr-unknown-role", Role: Role("viewer"), Status: StatusActive},
			want: false,
		},
		{
			name: "zero role cannot access owner routes",
			user: User{ID: "usr-zero-role", Status: StatusActive},
			want: false,
		},
		{
			name: "unknown status cannot access owner routes",
			user: User{ID: "usr-unknown-status", Role: RoleOwner, Status: Status("pending")},
			want: false,
		},
		{
			name: "zero status cannot access owner routes",
			user: User{ID: "usr-zero-status", Role: RoleOwner},
			want: false,
		},
		{
			name: "zero user cannot access owner routes",
			user: User{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanAccessOwner(tt.user); got != tt.want {
				t.Fatalf("CanAccessOwner(%+v) = %v, want %v", tt.user, got, tt.want)
			}
		})
	}
}

func TestCanAccessAdmin(t *testing.T) {
	tests := []struct {
		name string
		user User
		want bool
	}{
		{
			name: "active admin can access admin routes",
			user: User{ID: "usr-admin", Role: RoleAdmin, Status: StatusActive},
			want: true,
		},
		{
			name: "active owner cannot access admin routes",
			user: User{ID: "usr-owner", Role: RoleOwner, Status: StatusActive},
			want: false,
		},
		{
			name: "disabled admin cannot access admin routes",
			user: User{ID: "usr-disabled-admin", Role: RoleAdmin, Status: StatusDisabled},
			want: false,
		},
		{
			name: "unknown role cannot access admin routes",
			user: User{ID: "usr-unknown-role", Role: Role("viewer"), Status: StatusActive},
			want: false,
		},
		{
			name: "zero role cannot access admin routes",
			user: User{ID: "usr-zero-role", Status: StatusActive},
			want: false,
		},
		{
			name: "unknown status cannot access admin routes",
			user: User{ID: "usr-unknown-status", Role: RoleAdmin, Status: Status("pending")},
			want: false,
		},
		{
			name: "zero status cannot access admin routes",
			user: User{ID: "usr-zero-status", Role: RoleAdmin},
			want: false,
		},
		{
			name: "zero user cannot access admin routes",
			user: User{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanAccessAdmin(tt.user); got != tt.want {
				t.Fatalf("CanAccessAdmin(%+v) = %v, want %v", tt.user, got, tt.want)
			}
		})
	}
}
