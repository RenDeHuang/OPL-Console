package auth

import "testing"

func TestBootstrapUsersFromJSONDefaultsToBuiltInAdmin(t *testing.T) {
	users, err := BootstrapUsersFromJSON("")
	if err != nil {
		t.Fatalf("BootstrapUsersFromJSON: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("len(users) = %d", len(users))
	}
	if users[0].Email != "admin@opl.local" || users[0].Password != "OplAdminPass2026!" || users[0].Role != RoleAdmin {
		t.Fatalf("default user = %#v", users[0])
	}
}

func TestBootstrapUsersFromJSONParsesExplicitUsers(t *testing.T) {
	users, err := BootstrapUsersFromJSON(`[
		{"id":"usr-owner","email":"owner@example.com","password":"owner-pass","role":"owner","name":"Owner"},
		{"id":"usr-admin","email":"admin@example.com","password":"admin-pass","role":"admin","name":"Admin"}
	]`)
	if err != nil {
		t.Fatalf("BootstrapUsersFromJSON: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("len(users) = %d", len(users))
	}
	if users[0].Role != RoleOwner || users[1].Role != RoleAdmin {
		t.Fatalf("users = %#v", users)
	}
}

func TestBootstrapUsersFromJSONRejectsMissingRequiredFields(t *testing.T) {
	_, err := BootstrapUsersFromJSON(`[{"id":"usr-owner","email":"owner@example.com","role":"owner"}]`)
	if err == nil {
		t.Fatalf("expected error")
	}
}
