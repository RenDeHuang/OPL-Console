package store

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/RenDeHuang/opl-console/internal/console"
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

func TestGovernanceStoreBillingSupportPolicyApproval(t *testing.T) {
	ctx := context.Background()
	pool := authTestPool(ctx, t)
	store := NewGovernanceStore(pool)
	adminID := authTestID(t, "usr-admin")
	ownerID := authTestID(t, "usr-owner")
	orgID := authTestID(t, "org")
	billingID := authTestID(t, "billing")
	workspaceID := authTestID(t, "ws")
	policyID := authTestID(t, "policy")
	approvalID := authTestID(t, "approval")
	ledgerID := authTestID(t, "ledger")

	insertAuthUser(ctx, t, pool, adminID, "admin-governance@opl.local", "$2a$10$abcdefghijklmnopqrstuuqO4tLh7MZUn7CKiYQOz5PE4g5Q4v5k6", "admin", "active")
	insertAuthUser(ctx, t, pool, ownerID, "owner-governance@opl.local", "$2a$10$abcdefghijklmnopqrstuuqO4tLh7MZUn7CKiYQOz5PE4g5Q4v5k6", "owner", "active")
	_, err := pool.Exec(ctx, `
		INSERT INTO billing_accounts (id, owner_type, owner_id, balance_fen, frozen_fen, status)
		VALUES ($1, 'organization', $2, 5000, 1000, 'active')
	`, billingID, orgID)
	if err != nil {
		t.Fatalf("insert billing account: %v", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO organizations (id, name, billing_account_id, status)
		VALUES ($1, 'Governance Lab', $2, 'active')
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
		VALUES ($1, $2, 'Governance Workspace', 'basic', $3, 'running')
	`, workspaceID, billingID, workspaceID)
	if err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO billing_ledger_entries (id, billing_account_id, workspace_id, resource_type, resource_id, amount_fen, kind, description)
		VALUES ($1, $2, $3, 'workspace', $3, -1000, 'compute_hold', 'compute hold')
	`, ledgerID, billingID, workspaceID)
	if err != nil {
		t.Fatalf("insert ledger: %v", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO policies (id, organization_id, name, policy_type, rules, created_by_user_id)
		VALUES ($1, $2, 'Approval required', 'workspace_lifecycle', '{"requiresApproval":true}', $3)
	`, policyID, orgID, adminID)
	if err != nil {
		t.Fatalf("insert policy: %v", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO approvals (id, organization_id, policy_id, requester_user_id, action, object_type, object_id, status, reason)
		VALUES ($1, $2, $3, $4, 'workspace.create', 'workspace', $5, 'pending', 'policy required')
	`, approvalID, orgID, policyID, ownerID, workspaceID)
	if err != nil {
		t.Fatalf("insert approval: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM support_tickets WHERE billing_account_id = $1`, billingID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM approvals WHERE organization_id = $1`, orgID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM policies WHERE organization_id = $1`, orgID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM billing_ledger_entries WHERE billing_account_id = $1`, billingID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM workspaces WHERE billing_account_id = $1`, billingID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM memberships WHERE organization_id = $1`, orgID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM organizations WHERE id = $1`, orgID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM billing_accounts WHERE id = $1`, billingID)
	})

	wallet, err := store.WalletForUser(ctx, ownerID)
	if err != nil {
		t.Fatalf("WalletForUser: %v", err)
	}
	if wallet.AvailableFen != 4000 {
		t.Fatalf("wallet = %#v", wallet)
	}
	ledger, err := store.BillingLedgerForUser(ctx, ownerID)
	if err != nil {
		t.Fatalf("BillingLedgerForUser: %v", err)
	}
	if len(ledger) != 1 || ledger[0].ID != ledgerID {
		t.Fatalf("ledger = %#v", ledger)
	}
	ticket, err := store.CreateSupportTicket(ctx, ownerID, console.CreateSupportTicketRequest{
		WorkspaceID: workspaceID,
		Subject:     "Need help",
		Body:        "Workspace is blocked",
	})
	if err != nil {
		t.Fatalf("CreateSupportTicket: %v", err)
	}
	if ticket.Subject != "Need help" || ticket.Status != "open" {
		t.Fatalf("ticket = %#v", ticket)
	}
	tickets, err := store.SupportTicketsForUser(ctx, ownerID)
	if err != nil {
		t.Fatalf("SupportTicketsForUser: %v", err)
	}
	if len(tickets) != 1 || tickets[0].ID != ticket.ID {
		t.Fatalf("tickets = %#v", tickets)
	}
	policies, err := store.AdminPolicies(ctx)
	if err != nil {
		t.Fatalf("AdminPolicies: %v", err)
	}
	if len(policies) == 0 {
		t.Fatalf("policies = %#v", policies)
	}
	createdPolicy, err := store.CreatePolicy(ctx, adminID, console.CreatePolicyRequest{
		OrganizationID: orgID,
		Name:           "Quota guard",
		PolicyType:     "quota",
		Rules:          json.RawMessage(`{"maxWorkspaces":2}`),
	})
	if err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}
	if createdPolicy.PolicyType != "quota" {
		t.Fatalf("createdPolicy = %#v", createdPolicy)
	}
	approvals, err := store.AdminApprovals(ctx)
	if err != nil {
		t.Fatalf("AdminApprovals: %v", err)
	}
	if len(approvals) == 0 {
		t.Fatalf("approvals = %#v", approvals)
	}
	decision, err := store.DecideApproval(ctx, adminID, console.ApprovalDecisionRequest{
		ApprovalID:   approvalID,
		Decision:     "approved",
		DecisionNote: "approved for pilot",
	})
	if err != nil {
		t.Fatalf("DecideApproval: %v", err)
	}
	if decision.Status != "approved" || decision.ReviewerUserID != adminID {
		t.Fatalf("decision = %#v", decision)
	}
}
