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
		SELECT w.id, w.name, w.state, w.package_id,
		       COALESCE(c.status, ''), COALESCE(st.status, ''), COALESCE(a.status, ''),
		       COALESCE(t.status, ''),
		       COALESCE(route.metadata->>'workspaceUrl', ''),
		       COALESCE(route.provider, ''),
		       COALESCE(pkg.compute_hourly_fen * 24 * 7 + pkg.storage_gb_month_fen * pkg.storage_gb, 0)
		FROM workspaces w
		JOIN billing_accounts b ON b.id = w.billing_account_id
		JOIN organizations o ON o.billing_account_id = b.id
		JOIN memberships m ON m.organization_id = o.id
		JOIN workspace_packages pkg ON pkg.id = w.package_id
		LEFT JOIN compute_resources c ON c.id = w.compute_id
		LEFT JOIN storage_volumes st ON st.id = w.storage_id
		LEFT JOIN storage_attachments a ON a.id = w.attachment_id
		LEFT JOIN workspace_tokens t ON t.workspace_id = w.id AND t.status = 'active'
		LEFT JOIN managed_resource_views route ON route.workspace_id = w.id AND route.resource_type = 'route'
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
		if err := rows.Scan(
			&workspace.ID,
			&workspace.Name,
			&workspace.State,
			&workspace.PackageID,
			&workspace.ComputeStatus,
			&workspace.StorageStatus,
			&workspace.AttachmentStatus,
			&workspace.TokenStatus,
			&workspace.URL,
			&workspace.Provider,
			&workspace.EstimatedHoldFen,
		); err != nil {
			return nil, err
		}
		workspace.Policy = "managed"
		workspaces = append(workspaces, workspace)
	}
	return workspaces, rows.Err()
}

func (s *GovernanceStore) WorkspaceDetailForUser(ctx context.Context, userID string, workspaceID string) (console.WorkspaceDetail, error) {
	var detail console.WorkspaceDetail
	err := s.pool.QueryRow(ctx, `
		SELECT w.id, w.name, w.state, w.billing_account_id, w.package_id,
		       COALESCE(w.compute_id, ''), COALESCE(w.storage_id, ''), COALESCE(w.attachment_id, ''),
		       p.id, p.name, p.cpu, p.memory_gb, p.storage_gb, p.compute_hourly_fen, p.storage_gb_month_fen, p.available,
		       COALESCE(c.status, ''), COALESCE(st.status, ''), COALESCE(a.status, ''),
		       COALESCE(t.status, ''),
		       COALESCE(route.metadata->>'workspaceUrl', ''),
		       COALESCE(route.provider, '')
		FROM workspaces w
		JOIN billing_accounts b ON b.id = w.billing_account_id
		JOIN organizations o ON o.billing_account_id = b.id
		JOIN memberships m ON m.organization_id = o.id
		JOIN workspace_packages p ON p.id = w.package_id
		LEFT JOIN compute_resources c ON c.id = w.compute_id
		LEFT JOIN storage_volumes st ON st.id = w.storage_id
		LEFT JOIN storage_attachments a ON a.id = w.attachment_id
		LEFT JOIN workspace_tokens t ON t.workspace_id = w.id AND t.status = 'active'
		LEFT JOIN managed_resource_views route ON route.workspace_id = w.id AND route.resource_type = 'route'
		WHERE w.id = $1 AND m.user_id = $2 AND m.status = 'active'
		LIMIT 1
	`, workspaceID, userID).Scan(
		&detail.ID,
		&detail.Name,
		&detail.State,
		&detail.BillingAccountID,
		&detail.PackageID,
		&detail.ComputeID,
		&detail.StorageID,
		&detail.AttachmentID,
		&detail.Package.ID,
		&detail.Package.Name,
		&detail.Package.CPU,
		&detail.Package.MemoryGB,
		&detail.Package.StorageGB,
		&detail.Package.ComputeHourlyFen,
		&detail.Package.StorageGBMonthFen,
		&detail.Package.Available,
		&detail.ComputeStatus,
		&detail.StorageStatus,
		&detail.AttachmentStatus,
		&detail.TokenStatus,
		&detail.URL,
		&detail.Provider,
	)
	if err != nil {
		return console.WorkspaceDetail{}, err
	}
	detail.Policy = "managed"
	detail.EstimatedHoldFen = detail.Package.ComputeHourlyFen*24*7 + detail.Package.StorageGBMonthFen*int64(detail.Package.StorageGB)
	var loadErr error
	if detail.LifecycleSteps, loadErr = s.workspaceLifecycleSteps(ctx, workspaceID); loadErr != nil {
		return console.WorkspaceDetail{}, loadErr
	}
	if detail.LedgerEntries, loadErr = s.billingLedgerForWorkspace(ctx, detail.BillingAccountID, workspaceID); loadErr != nil {
		return console.WorkspaceDetail{}, loadErr
	}
	if detail.Receipts, loadErr = s.receiptsForSubject(ctx, workspaceID); loadErr != nil {
		return console.WorkspaceDetail{}, loadErr
	}
	if detail.SupportTickets, loadErr = s.supportTicketsForWorkspace(ctx, detail.BillingAccountID, workspaceID); loadErr != nil {
		return console.WorkspaceDetail{}, loadErr
	}
	if detail.AuditEvents, loadErr = s.auditEventsForObject(ctx, workspaceID); loadErr != nil {
		return console.WorkspaceDetail{}, loadErr
	}
	return detail, nil
}

func (s *GovernanceStore) workspaceLifecycleSteps(ctx context.Context, workspaceID string) ([]console.LifecycleStepView, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT step_name, desired_state, actual_state, provider_resource_id, error_code, last_checked_at
		FROM workspace_lifecycle_steps
		WHERE workspace_id = $1
		ORDER BY created_at ASC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var steps []console.LifecycleStepView
	for rows.Next() {
		var step console.LifecycleStepView
		var lastCheckedAt time.Time
		if err := rows.Scan(
			&step.StepName,
			&step.DesiredState,
			&step.ActualState,
			&step.ProviderResourceID,
			&step.ErrorCode,
			&lastCheckedAt,
		); err != nil {
			return nil, err
		}
		step.LastCheckedAt = lastCheckedAt.Format(time.RFC3339)
		steps = append(steps, step)
	}
	return steps, rows.Err()
}

func (s *GovernanceStore) receiptsForSubject(ctx context.Context, subjectID string) ([]console.ReceiptView, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, receipt_type, subject_type, subject_id, COALESCE(operation_id, ''), payload
		FROM receipts
		WHERE subject_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var receipts []console.ReceiptView
	for rows.Next() {
		var receipt console.ReceiptView
		if err := rows.Scan(
			&receipt.ID,
			&receipt.ReceiptType,
			&receipt.SubjectType,
			&receipt.SubjectID,
			&receipt.OperationID,
			&receipt.Payload,
		); err != nil {
			return nil, err
		}
		receipts = append(receipts, receipt)
	}
	return receipts, rows.Err()
}

func (s *GovernanceStore) auditEventsForObject(ctx context.Context, objectID string) ([]console.AuditEventView, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, actor_user_id, action, object_type, object_id, result, metadata, created_at
		FROM audit_events
		WHERE object_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, objectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []console.AuditEventView
	for rows.Next() {
		var event console.AuditEventView
		var createdAt time.Time
		if err := rows.Scan(
			&event.ID,
			&event.ActorUserID,
			&event.Action,
			&event.ObjectType,
			&event.ObjectID,
			&event.Result,
			&event.Metadata,
			&createdAt,
		); err != nil {
			return nil, err
		}
		event.CreatedAt = createdAt.Format(time.RFC3339)
		events = append(events, event)
	}
	return events, rows.Err()
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

func (s *GovernanceStore) AdminOrganizations(ctx context.Context) ([]console.OrganizationView, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, status
		FROM organizations
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var organizations []console.OrganizationView
	for rows.Next() {
		var organization console.OrganizationView
		if err := rows.Scan(&organization.ID, &organization.Name, &organization.Status); err != nil {
			return nil, err
		}
		organizations = append(organizations, organization)
	}
	return organizations, rows.Err()
}

func (s *GovernanceStore) AdminTeams(ctx context.Context) ([]console.TeamView, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, name, status
		FROM teams
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []console.TeamView
	for rows.Next() {
		var team console.TeamView
		if err := rows.Scan(&team.ID, &team.OrganizationID, &team.Name, &team.Status); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	return teams, rows.Err()
}

func (s *GovernanceStore) AdminRoles(ctx context.Context) ([]console.RoleView, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, COALESCE(organization_id, ''), name, scope, permissions
		FROM roles
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []console.RoleView
	for rows.Next() {
		var role console.RoleView
		if err := rows.Scan(&role.ID, &role.OrganizationID, &role.Name, &role.Scope, &role.Permissions); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (s *GovernanceStore) AdminManagedResources(ctx context.Context) ([]console.ManagedResourceView, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, resource_type, resource_id, display_name,
		       provider, status, policy_state, COALESCE(workspace_id, ''), metadata
		FROM managed_resource_views
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []console.ManagedResourceView
	for rows.Next() {
		var resource console.ManagedResourceView
		if err := rows.Scan(
			&resource.ID,
			&resource.OrganizationID,
			&resource.ResourceType,
			&resource.ResourceID,
			&resource.DisplayName,
			&resource.Provider,
			&resource.Status,
			&resource.PolicyState,
			&resource.WorkspaceID,
			&resource.Metadata,
		); err != nil {
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, rows.Err()
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

func (s *GovernanceStore) AdminBillingLedger(ctx context.Context) ([]console.BillingLedgerEntryView, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT e.id, COALESCE(e.workspace_id, ''), e.resource_type, COALESCE(e.resource_id, ''),
		       e.amount_fen, e.kind, e.description, e.created_at
		FROM billing_ledger_entries e
		ORDER BY e.created_at DESC
		LIMIT 200
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBillingLedgerRows(rows)
}

func (s *GovernanceStore) billingLedgerForWorkspace(ctx context.Context, billingAccountID string, workspaceID string) ([]console.BillingLedgerEntryView, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT e.id, COALESCE(e.workspace_id, ''), e.resource_type, COALESCE(e.resource_id, ''),
		       e.amount_fen, e.kind, e.description, e.created_at
		FROM billing_ledger_entries e
		WHERE e.billing_account_id = $1 AND (e.workspace_id = $2 OR e.resource_id = $2)
		ORDER BY e.created_at DESC
		LIMIT 50
	`, billingAccountID, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBillingLedgerRows(rows)
}

type billingLedgerRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanBillingLedgerRows(rows billingLedgerRows) ([]console.BillingLedgerEntryView, error) {
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
		SELECT t.id, COALESCE(t.workspace_id, ''), t.subject, t.body, t.status,
		       t.priority, COALESCE(t.assignee_user_id, ''), t.failed_lifecycle_step,
		       t.fabric_error_code, t.runtime_status, t.ledger_summary, t.created_at
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
	return scanSupportTicketRows(rows)
}

func (s *GovernanceStore) AdminSupportTickets(ctx context.Context) ([]console.SupportTicketView, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t.id, COALESCE(t.workspace_id, ''), t.subject, t.body, t.status,
		       t.priority, COALESCE(t.assignee_user_id, ''), t.failed_lifecycle_step,
		       t.fabric_error_code, t.runtime_status, t.ledger_summary, t.created_at
		FROM support_tickets t
		ORDER BY t.created_at DESC
		LIMIT 200
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSupportTicketRows(rows)
}

func (s *GovernanceStore) supportTicketsForWorkspace(ctx context.Context, billingAccountID string, workspaceID string) ([]console.SupportTicketView, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t.id, COALESCE(t.workspace_id, ''), t.subject, t.body, t.status,
		       t.priority, COALESCE(t.assignee_user_id, ''), t.failed_lifecycle_step,
		       t.fabric_error_code, t.runtime_status, t.ledger_summary, t.created_at
		FROM support_tickets t
		WHERE t.billing_account_id = $1 AND t.workspace_id = $2
		ORDER BY t.created_at DESC
	`, billingAccountID, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSupportTicketRows(rows)
}

type supportTicketRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanSupportTicketRows(rows supportTicketRows) ([]console.SupportTicketView, error) {
	var tickets []console.SupportTicketView
	for rows.Next() {
		var ticket console.SupportTicketView
		var createdAt time.Time
		if err := rows.Scan(
			&ticket.ID,
			&ticket.WorkspaceID,
			&ticket.Subject,
			&ticket.Body,
			&ticket.Status,
			&ticket.Priority,
			&ticket.AssigneeUserID,
			&ticket.FailedLifecycleStep,
			&ticket.FabricErrorCode,
			&ticket.RuntimeStatus,
			&ticket.LedgerSummary,
			&createdAt,
		); err != nil {
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
		RETURNING id, COALESCE(workspace_id, ''), subject, body, status, priority,
		          COALESCE(assignee_user_id, ''), failed_lifecycle_step, fabric_error_code,
		          runtime_status, ledger_summary, created_at
	`, userID, id, request.WorkspaceID, request.Subject, request.Body).Scan(
		&ticket.ID,
		&ticket.WorkspaceID,
		&ticket.Subject,
		&ticket.Body,
		&ticket.Status,
		&ticket.Priority,
		&ticket.AssigneeUserID,
		&ticket.FailedLifecycleStep,
		&ticket.FabricErrorCode,
		&ticket.RuntimeStatus,
		&ticket.LedgerSummary,
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
		       status, reason, decision_note, context
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
			&approval.Context,
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
		          status, reason, decision_note, context
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
		&approval.Context,
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
