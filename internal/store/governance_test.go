package store

import (
	"context"
	"testing"
)

func TestGovernanceStoreReadModel(t *testing.T) {
	ctx := context.Background()
	pool := authTestPool(ctx, t)
	store := NewGovernanceStore(pool)
	adminID := authTestID(t, "usr-admin")
	ownerID := authTestID(t, "usr-owner")
	orgID := authTestID(t, "org")
	billingID := authTestID(t, "billing")
	workspaceID := authTestID(t, "ws")

	insertAuthUser(ctx, t, pool, adminID, "admin-readmodel@opl.local", "$2a$10$abcdefghijklmnopqrstuuqO4tLh7MZUn7CKiYQOz5PE4g5Q4v5k6", "admin", "active")
	insertAuthUser(ctx, t, pool, ownerID, "owner-readmodel@opl.local", "$2a$10$abcdefghijklmnopqrstuuqO4tLh7MZUn7CKiYQOz5PE4g5Q4v5k6", "owner", "active")
	_, err := pool.Exec(ctx, `
		INSERT INTO billing_accounts (id, owner_type, owner_id, balance_fen, frozen_fen, status)
		VALUES ($1, 'organization', $2, 10000, 0, 'active')
	`, billingID, orgID)
	if err != nil {
		t.Fatalf("insert billing account: %v", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO organizations (id, name, billing_account_id, status)
		VALUES ($1, 'Read Model Lab', $2, 'active')
	`, orgID, billingID)
	if err != nil {
		t.Fatalf("insert organization: %v", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO memberships (id, organization_id, user_id, role, status)
		VALUES ($1, $2, $3, 'owner', 'active')
	`, authTestID(t, "member"), orgID, ownerID)
	if err != nil {
		t.Fatalf("insert membership: %v", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspaces (id, billing_account_id, name, package_id, slug, state)
		VALUES ($1, $2, 'Read Model Workspace', 'basic', $3, 'running')
	`, workspaceID, billingID, workspaceID)
	if err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM workspaces WHERE id = $1`, workspaceID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM memberships WHERE organization_id = $1`, orgID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM organizations WHERE id = $1`, orgID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM billing_accounts WHERE id = $1`, billingID)
	})

	user, err := store.UserByID(ctx, ownerID)
	if err != nil {
		t.Fatalf("UserByID: %v", err)
	}
	if user.Email != "owner-readmodel@opl.local" {
		t.Fatalf("user = %#v", user)
	}
	org, err := store.PrimaryOrganizationForUser(ctx, ownerID)
	if err != nil {
		t.Fatalf("PrimaryOrganizationForUser: %v", err)
	}
	if org.ID != orgID {
		t.Fatalf("organization = %#v", org)
	}
	packages, err := store.Packages(ctx)
	if err != nil {
		t.Fatalf("Packages: %v", err)
	}
	if len(packages) < 2 {
		t.Fatalf("packages = %#v", packages)
	}
	workspaces, err := store.WorkspacesForUser(ctx, ownerID)
	if err != nil {
		t.Fatalf("WorkspacesForUser: %v", err)
	}
	if len(workspaces) != 1 || workspaces[0].ID != workspaceID {
		t.Fatalf("workspaces = %#v", workspaces)
	}
	users, err := store.AdminUsers(ctx)
	if err != nil {
		t.Fatalf("AdminUsers: %v", err)
	}
	if len(users) < 2 {
		t.Fatalf("users = %#v", users)
	}
}
