package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/RenDeHuang/opl-console/internal/console"
)

type GovernanceStore struct {
	pool *pgxpool.Pool
}

var _ console.Repository = (*GovernanceStore)(nil)

func NewGovernanceStore(pool *pgxpool.Pool) *GovernanceStore {
	return &GovernanceStore{pool: pool}
}

func (s *GovernanceStore) UserByID(ctx context.Context, userID string) (console.UserView, error) {
	var user console.UserView
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, role, status
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Email, &user.Role, &user.Status)
	return user, err
}

func (s *GovernanceStore) PrimaryOrganizationForUser(ctx context.Context, userID string) (console.OrganizationView, error) {
	var organization console.OrganizationView
	err := s.pool.QueryRow(ctx, `
		SELECT o.id, o.name, o.status
		FROM organizations o
		JOIN memberships m ON m.organization_id = o.id
		WHERE m.user_id = $1 AND m.status = 'active'
		ORDER BY o.created_at ASC
		LIMIT 1
	`, userID).Scan(&organization.ID, &organization.Name, &organization.Status)
	return organization, err
}

func (s *GovernanceStore) Packages(ctx context.Context) ([]console.Package, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, cpu, memory_gb, storage_gb, compute_hourly_fen, storage_gb_month_fen, available
		FROM workspace_packages
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []console.Package
	for rows.Next() {
		var pkg console.Package
		if err := rows.Scan(
			&pkg.ID,
			&pkg.Name,
			&pkg.CPU,
			&pkg.MemoryGB,
			&pkg.StorageGB,
			&pkg.ComputeHourlyFen,
			&pkg.StorageGBMonthFen,
			&pkg.Available,
		); err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}
	return packages, rows.Err()
}

func (s *GovernanceStore) WorkspacesForUser(ctx context.Context, userID string) ([]console.ManagedWorkspace, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT w.id, w.name, w.state
		FROM workspaces w
		JOIN billing_accounts b ON b.id = w.billing_account_id
		JOIN organizations o ON o.billing_account_id = b.id
		JOIN memberships m ON m.organization_id = o.id
		WHERE m.user_id = $1 AND m.status = 'active'
		ORDER BY w.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []console.ManagedWorkspace
	for rows.Next() {
		var workspace console.ManagedWorkspace
		if err := rows.Scan(&workspace.ID, &workspace.Name, &workspace.State); err != nil {
			return nil, err
		}
		workspace.Policy = "managed"
		workspaces = append(workspaces, workspace)
	}
	return workspaces, rows.Err()
}

func (s *GovernanceStore) AdminUsers(ctx context.Context) ([]console.UserView, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, email, role, status
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []console.UserView
	for rows.Next() {
		var user console.UserView
		if err := rows.Scan(&user.ID, &user.Email, &user.Role, &user.Status); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *GovernanceStore) WalletForUser(ctx context.Context, userID string) (console.WalletView, error) {
	var wallet console.WalletView
	err := s.pool.QueryRow(ctx, `
		SELECT b.id, b.balance_fen, b.frozen_fen, b.balance_fen - b.frozen_fen
		FROM billing_accounts b
		JOIN organizations o ON o.billing_account_id = b.id
		JOIN memberships m ON m.organization_id = o.id
		WHERE m.user_id = $1 AND m.status = 'active'
		ORDER BY o.created_at ASC
		LIMIT 1
	`, userID).Scan(&wallet.BillingAccountID, &wallet.BalanceFen, &wallet.FrozenFen, &wallet.AvailableFen)
	return wallet, err
}

func (s *GovernanceStore) BillingLedgerForUser(ctx context.Context, userID string) ([]console.BillingLedgerEntryView, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT e.id, COALESCE(e.workspace_id, ''), e.resource_type, COALESCE(e.resource_id, ''),
		       e.amount_fen, e.kind, e.description, e.created_at
		FROM billing_ledger_entries e
		JOIN billing_accounts b ON b.id = e.billing_account_id
		JOIN organizations o ON o.billing_account_id = b.id
		JOIN memberships m ON m.organization_id = o.id
		WHERE m.user_id = $1 AND m.status = 'active'
		ORDER BY e.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []console.BillingLedgerEntryView
	for rows.Next() {
		var entry console.BillingLedgerEntryView
		var createdAt time.Time
		if err := rows.Scan(
			&entry.ID,
			&entry.WorkspaceID,
			&entry.ResourceType,
			&entry.ResourceID,
			&entry.AmountFen,
			&entry.Kind,
			&entry.Description,
			&createdAt,
		); err != nil {
			return nil, err
		}
		entry.CreatedAt = createdAt.Format(time.RFC3339)
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (s *GovernanceStore) SupportTicketsForUser(ctx context.Context, userID string) ([]console.SupportTicketView, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t.id, COALESCE(t.workspace_id, ''), t.subject, t.body, t.status, t.created_at
		FROM support_tickets t
		JOIN billing_accounts b ON b.id = t.billing_account_id
		JOIN organizations o ON o.billing_account_id = b.id
		JOIN memberships m ON m.organization_id = o.id
		WHERE m.user_id = $1 AND m.status = 'active'
		ORDER BY t.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []console.SupportTicketView
	for rows.Next() {
		var ticket console.SupportTicketView
		var createdAt time.Time
		if err := rows.Scan(&ticket.ID, &ticket.WorkspaceID, &ticket.Subject, &ticket.Body, &ticket.Status, &createdAt); err != nil {
			return nil, err
		}
		ticket.CreatedAt = createdAt.Format(time.RFC3339)
		tickets = append(tickets, ticket)
	}
	return tickets, rows.Err()
}

func (s *GovernanceStore) CreateSupportTicket(ctx context.Context, userID string, request console.CreateSupportTicketRequest) (console.SupportTicketView, error) {
	id, err := randomStoreID("ticket")
	if err != nil {
		return console.SupportTicketView{}, err
	}
	var ticket console.SupportTicketView
	var createdAt time.Time
	err = s.pool.QueryRow(ctx, `
		WITH primary_org AS (
			SELECT o.billing_account_id
			FROM organizations o
			JOIN memberships m ON m.organization_id = o.id
			WHERE m.user_id = $1 AND m.status = 'active'
			ORDER BY o.created_at ASC
			LIMIT 1
		)
		INSERT INTO support_tickets (id, billing_account_id, user_id, workspace_id, subject, body, status)
		SELECT $2, billing_account_id, $1, NULLIF($3, ''), $4, $5, 'open'
		FROM primary_org
		RETURNING id, COALESCE(workspace_id, ''), subject, body, status, created_at
	`, userID, id, request.WorkspaceID, request.Subject, request.Body).Scan(
		&ticket.ID,
		&ticket.WorkspaceID,
		&ticket.Subject,
		&ticket.Body,
		&ticket.Status,
		&createdAt,
	)
	if err != nil {
		return console.SupportTicketView{}, err
	}
	ticket.CreatedAt = createdAt.Format(time.RFC3339)
	return ticket, nil
}

func (s *GovernanceStore) AdminPolicies(ctx context.Context) ([]console.PolicyView, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, name, policy_type, status, rules
		FROM policies
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []console.PolicyView
	for rows.Next() {
		var policy console.PolicyView
		if err := rows.Scan(&policy.ID, &policy.OrganizationID, &policy.Name, &policy.PolicyType, &policy.Status, &policy.Rules); err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}
	return policies, rows.Err()
}

func (s *GovernanceStore) CreatePolicy(ctx context.Context, actorUserID string, request console.CreatePolicyRequest) (console.PolicyView, error) {
	id, err := randomStoreID("policy")
	if err != nil {
		return console.PolicyView{}, err
	}
	if len(request.Rules) == 0 {
		request.Rules = []byte(`{}`)
	}
	var policy console.PolicyView
	err = s.pool.QueryRow(ctx, `
		INSERT INTO policies (id, organization_id, name, policy_type, status, rules, created_by_user_id)
		VALUES ($1, $2, $3, $4, 'active', $5, $6)
		RETURNING id, organization_id, name, policy_type, status, rules
	`, id, request.OrganizationID, request.Name, request.PolicyType, request.Rules, actorUserID).Scan(
		&policy.ID,
		&policy.OrganizationID,
		&policy.Name,
		&policy.PolicyType,
		&policy.Status,
		&policy.Rules,
	)
	return policy, err
}

func (s *GovernanceStore) AdminApprovals(ctx context.Context) ([]console.ApprovalView, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, COALESCE(policy_id, ''), requester_user_id,
		       COALESCE(reviewer_user_id, ''), action, object_type, object_id,
		       status, reason, decision_note
		FROM approvals
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var approvals []console.ApprovalView
	for rows.Next() {
		var approval console.ApprovalView
		if err := rows.Scan(
			&approval.ID,
			&approval.OrganizationID,
			&approval.PolicyID,
			&approval.RequesterUserID,
			&approval.ReviewerUserID,
			&approval.Action,
			&approval.ObjectType,
			&approval.ObjectID,
			&approval.Status,
			&approval.Reason,
			&approval.DecisionNote,
		); err != nil {
			return nil, err
		}
		approvals = append(approvals, approval)
	}
	return approvals, rows.Err()
}

func (s *GovernanceStore) DecideApproval(ctx context.Context, actorUserID string, request console.ApprovalDecisionRequest) (console.ApprovalView, error) {
	if request.Decision != "approved" && request.Decision != "rejected" {
		return console.ApprovalView{}, fmt.Errorf("unsupported approval decision: %s", request.Decision)
	}
	var approval console.ApprovalView
	err := s.pool.QueryRow(ctx, `
		UPDATE approvals
		SET status = $1,
		    reviewer_user_id = $2,
		    decision_note = $3,
		    decided_at = now()
		WHERE id = $4 AND status = 'pending'
		RETURNING id, organization_id, COALESCE(policy_id, ''), requester_user_id,
		          COALESCE(reviewer_user_id, ''), action, object_type, object_id,
		          status, reason, decision_note
	`, request.Decision, actorUserID, request.DecisionNote, request.ApprovalID).Scan(
		&approval.ID,
		&approval.OrganizationID,
		&approval.PolicyID,
		&approval.RequesterUserID,
		&approval.ReviewerUserID,
		&approval.Action,
		&approval.ObjectType,
		&approval.ObjectID,
		&approval.Status,
		&approval.Reason,
		&approval.DecisionNote,
	)
	return approval, err
}

func randomStoreID(prefix string) (string, error) {
	raw := make([]byte, 8)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return prefix + "-" + hex.EncodeToString(raw), nil
}
