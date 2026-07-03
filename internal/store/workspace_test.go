package store

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/RenDeHuang/opl-console/internal/workspace"
)

func TestWorkspaceStorePrepareCreateRequiresMembershipAndDetectsPolicy(t *testing.T) {
	ctx := context.Background()
	pool := authTestPool(ctx, t)
	store := NewWorkspaceStore(pool)
	ownerID := authTestID(t, "usr-owner")
	adminID := authTestID(t, "usr-admin")
	orgID := authTestID(t, "org")
	billingID := authTestID(t, "billing")
	policyID := authTestID(t, "policy")

	insertAuthUser(ctx, t, pool, ownerID, "owner-workspace-store@opl.local", "$2a$10$abcdefghijklmnopqrstuuqO4tLh7MZUn7CKiYQOz5PE4g5Q4v5k6", "owner", "active")
	insertAuthUser(ctx, t, pool, adminID, "admin-workspace-store@opl.local", "$2a$10$abcdefghijklmnopqrstuuqO4tLh7MZUn7CKiYQOz5PE4g5Q4v5k6", "admin", "active")
	insertWorkspaceOrg(ctx, t, pool, orgID, billingID, ownerID, 10000)
	_, err := pool.Exec(ctx, `
		INSERT INTO policies (id, organization_id, name, policy_type, rules, created_by_user_id)
		VALUES ($1, $2, 'Workspace approval', 'workspace_lifecycle', '{"requiresApproval":true}', $3)
	`, policyID, orgID, adminID)
	if err != nil {
		t.Fatalf("insert policy: %v", err)
	}

	result, err := store.PrepareCreate(ctx, workspace.CreateWorkspaceRequest{
		WorkspaceID:      "ws-alpha",
		BillingAccountID: billingID,
		PackageID:        "basic",
		ActorUserID:      ownerID,
	})
	if err != nil {
		t.Fatalf("PrepareCreate: %v", err)
	}
	if result.OrganizationID != orgID || !result.RequiresApproval || result.Package.ID != "basic" || result.HoldAmountFen <= 0 {
		t.Fatalf("result = %#v", result)
	}

	_, err = store.PrepareCreate(ctx, workspace.CreateWorkspaceRequest{
		WorkspaceID:      "ws-alpha",
		BillingAccountID: billingID,
		PackageID:        "basic",
		ActorUserID:      "missing-user",
	})
	if err == nil {
		t.Fatal("PrepareCreate returned nil error for non-member actor")
	}
}

func TestWorkspaceStoreSaveCreatedPersistsFacadeState(t *testing.T) {
	ctx := context.Background()
	pool := authTestPool(ctx, t)
	store := NewWorkspaceStore(pool)
	ownerID := authTestID(t, "usr-owner")
	orgID := authTestID(t, "org")
	billingID := authTestID(t, "billing")
	workspaceID := authTestID(t, "ws")

	insertAuthUser(ctx, t, pool, ownerID, "owner-workspace-save@opl.local", "$2a$10$abcdefghijklmnopqrstuuqO4tLh7MZUn7CKiYQOz5PE4g5Q4v5k6", "owner", "active")
	insertWorkspaceOrg(ctx, t, pool, orgID, billingID, ownerID, 10000)

	err := store.SaveCreated(ctx, workspace.CreatedWorkspace{
		WorkspaceID:       workspaceID,
		Name:              "Saved Workspace",
		BillingAccountID:  billingID,
		OrganizationID:    orgID,
		PackageID:         "basic",
		ComputeID:         "cmp-" + workspaceID,
		StorageID:         "stg-" + workspaceID,
		AttachmentID:      "att-" + workspaceID,
		ComputeProviderID: "local-compute/" + workspaceID,
		StorageProviderID: "local-storage/" + workspaceID,
		AttachProviderID:  "local-attach/" + workspaceID,
		RouteProviderID:   "local-route/" + workspaceID,
		RouteURL:          "https://workspace.example/" + workspaceID,
		Token:             "share-token",
	})
	if err != nil {
		t.Fatalf("SaveCreated: %v", err)
	}

	var state string
	if err := pool.QueryRow(ctx, `SELECT state FROM workspaces WHERE id = $1`, workspaceID).Scan(&state); err != nil {
		t.Fatalf("query workspace: %v", err)
	}
	if state != "running" {
		t.Fatalf("state = %q, want running", state)
	}
	var viewCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM managed_resource_views WHERE workspace_id = $1`, workspaceID).Scan(&viewCount); err != nil {
		t.Fatalf("query managed_resource_views: %v", err)
	}
	if viewCount != 4 {
		t.Fatalf("managed resource view count = %d, want 4", viewCount)
	}
	var tokenStatus string
	if err := pool.QueryRow(ctx, `SELECT status FROM workspace_tokens WHERE workspace_id = $1`, workspaceID).Scan(&tokenStatus); err != nil {
		t.Fatalf("query workspace token: %v", err)
	}
	if tokenStatus != "active" {
		t.Fatalf("token status = %q, want active", tokenStatus)
	}
	handoff, err := store.Handoff(ctx, workspace.HandoffRequest{WorkspaceID: workspaceID, Token: "share-token"})
	if err != nil {
		t.Fatalf("Handoff: %v", err)
	}
	if handoff.URL != "https://workspace.example/"+workspaceID || handoff.State != "running" {
		t.Fatalf("handoff = %#v", handoff)
	}
	runtime, err := store.RuntimeForAction(ctx, workspace.ActionRequest{WorkspaceID: workspaceID, ActorUserID: ownerID})
	if err != nil {
		t.Fatalf("RuntimeForAction: %v", err)
	}
	if runtime.ComputeID != "cmp-"+workspaceID || runtime.StorageID != "stg-"+workspaceID {
		t.Fatalf("runtime = %#v", runtime)
	}
	if err := store.UpdateWorkspaceState(ctx, workspace.StateChange{WorkspaceID: workspaceID, ActorUserID: ownerID, State: "configured"}); err != nil {
		t.Fatalf("UpdateWorkspaceState: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT state FROM workspaces WHERE id = $1`, workspaceID).Scan(&state); err != nil {
		t.Fatalf("query configured workspace: %v", err)
	}
	if state != "configured" {
		t.Fatalf("state = %q, want configured", state)
	}
	if err := store.ReplaceActiveToken(ctx, workspace.TokenChange{
		WorkspaceID: workspaceID,
		ActorUserID: ownerID,
		Token:       "new-token",
		URL:         "https://workspace.example/" + workspaceID + "?token=new-token",
	}); err != nil {
		t.Fatalf("ReplaceActiveToken: %v", err)
	}
	if _, err := store.Handoff(ctx, workspace.HandoffRequest{WorkspaceID: workspaceID, Token: "share-token"}); err == nil {
		t.Fatal("Handoff succeeded with replaced old token")
	}
	if _, err := store.Handoff(ctx, workspace.HandoffRequest{WorkspaceID: workspaceID, Token: "new-token"}); err != nil {
		t.Fatalf("Handoff new token: %v", err)
	}
	if err := store.DeleteActiveToken(ctx, workspace.ActionRequest{WorkspaceID: workspaceID, ActorUserID: ownerID}); err != nil {
		t.Fatalf("DeleteActiveToken: %v", err)
	}
	if _, err := store.Handoff(ctx, workspace.HandoffRequest{WorkspaceID: workspaceID, Token: "new-token"}); err == nil {
		t.Fatal("Handoff succeeded after token delete")
	}
}

func insertWorkspaceOrg(ctx context.Context, t *testing.T, pool *pgxpool.Pool, orgID string, billingID string, ownerID string, balanceFen int64) {
	t.Helper()
	_, err := pool.Exec(ctx, `
		INSERT INTO billing_accounts (id, owner_type, owner_id, balance_fen, frozen_fen, status)
		VALUES ($1, 'organization', $2, $3, 0, 'active')
	`, billingID, orgID, balanceFen)
	if err != nil {
		t.Fatalf("insert billing account: %v", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO organizations (id, name, billing_account_id, status)
		VALUES ($1, 'Workspace Store Lab', $2, 'active')
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
	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM managed_resource_views WHERE organization_id = $1`, orgID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM workspace_tokens WHERE workspace_id IN (SELECT id FROM workspaces WHERE billing_account_id = $1)`, billingID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM workspaces WHERE billing_account_id = $1`, billingID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM storage_attachments WHERE compute_id IN (SELECT id FROM compute_resources WHERE billing_account_id = $1)`, billingID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM compute_resources WHERE billing_account_id = $1`, billingID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM storage_volumes WHERE billing_account_id = $1`, billingID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM approvals WHERE organization_id = $1`, orgID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM policies WHERE organization_id = $1`, orgID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM memberships WHERE organization_id = $1`, orgID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM organizations WHERE id = $1`, orgID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM billing_accounts WHERE id = $1`, billingID)
	})
}
