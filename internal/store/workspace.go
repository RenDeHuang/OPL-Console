package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/RenDeHuang/opl-console/internal/auth"
	"github.com/RenDeHuang/opl-console/internal/workspace"
)

type WorkspaceStore struct {
	pool *pgxpool.Pool
}

func NewWorkspaceStore(pool *pgxpool.Pool) *WorkspaceStore {
	return &WorkspaceStore{pool: pool}
}

func (s *WorkspaceStore) PrepareCreate(ctx context.Context, request workspace.CreateWorkspaceRequest) (workspace.CreateContext, error) {
	var result workspace.CreateContext
	err := s.pool.QueryRow(ctx, `
		SELECT o.id
		FROM organizations o
		JOIN memberships m ON m.organization_id = o.id
		WHERE o.billing_account_id = $1
		  AND m.user_id = $2
		  AND m.status = 'active'
		  AND o.status = 'active'
		LIMIT 1
	`, request.BillingAccountID, request.ActorUserID).Scan(&result.OrganizationID)
	if err != nil {
		return workspace.CreateContext{}, err
	}
	result.ActorUserID = request.ActorUserID
	var computeHourlyFen int64
	var storageGBMonthFen int64
	err = s.pool.QueryRow(ctx, `
		SELECT id, cpu, memory_gb, storage_gb, compute_hourly_fen, storage_gb_month_fen
		FROM workspace_packages
		WHERE id = $1 AND available = true
	`, request.PackageID).Scan(
		&result.Package.ID,
		&result.Package.CPU,
		&result.Package.MemoryGB,
		&result.Package.StorageGB,
		&computeHourlyFen,
		&storageGBMonthFen,
	)
	if err != nil {
		return workspace.CreateContext{}, err
	}
	result.HoldAmountFen = computeHourlyFen*24 + storageGBMonthFen*int64(result.Package.StorageGB)
	result.RequiresApproval, err = s.workspaceCreateRequiresApproval(ctx, result.OrganizationID)
	if err != nil {
		return workspace.CreateContext{}, err
	}
	return result, nil
}

func (s *WorkspaceStore) workspaceCreateRequiresApproval(ctx context.Context, organizationID string) (bool, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT rules
		FROM policies
		WHERE organization_id = $1
		  AND policy_type = 'workspace_lifecycle'
		  AND status = 'active'
	`, organizationID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var raw json.RawMessage
		if err := rows.Scan(&raw); err != nil {
			return false, err
		}
		var rules struct {
			RequiresApproval bool `json:"requiresApproval"`
		}
		if err := json.Unmarshal(raw, &rules); err != nil {
			return false, err
		}
		if rules.RequiresApproval {
			return true, nil
		}
	}
	return false, rows.Err()
}

func (s *WorkspaceStore) CreateApproval(ctx context.Context, request workspace.ApprovalRequest) (string, error) {
	id, err := randomStoreID("approval")
	if err != nil {
		return "", err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO approvals (id, organization_id, policy_id, requester_user_id, action, object_type, object_id, status, reason)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7, 'pending', $8)
	`, id, request.OrganizationID, request.PolicyID, request.RequesterUserID, request.Action, request.ObjectType, request.ObjectID, request.Reason)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *WorkspaceStore) SaveCreated(ctx context.Context, record workspace.CreatedWorkspace) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO compute_resources (id, billing_account_id, package_id, provider_resource_id, status)
		VALUES ($1, $2, $3, $4, 'running')
	`, record.ComputeID, record.BillingAccountID, record.PackageID, record.ComputeProviderID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO storage_volumes (id, billing_account_id, package_id, provider_resource_id, size_gb, status)
		SELECT $1, $2, $3, $4, storage_gb, 'available'
		FROM workspace_packages
		WHERE id = $3
	`, record.StorageID, record.BillingAccountID, record.PackageID, record.StorageProviderID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO storage_attachments (id, compute_id, storage_id, mount_path, status)
		VALUES ($1, $2, $3, '/data', 'attached')
	`, record.AttachmentID, record.ComputeID, record.StorageID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO workspaces (id, billing_account_id, name, package_id, compute_id, storage_id, attachment_id, slug, state)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $1, 'running')
	`, record.WorkspaceID, record.BillingAccountID, record.Name, record.PackageID, record.ComputeID, record.StorageID, record.AttachmentID); err != nil {
		return err
	}
	tokenID, err := randomStoreID("token")
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO workspace_tokens (id, workspace_id, token_hash, status)
		VALUES ($1, $2, $3, 'active')
	`, tokenID, record.WorkspaceID, auth.HashToken(record.Token)); err != nil {
		return err
	}
	views := []struct {
		resourceType string
		resourceID   string
		displayName  string
		providerID   string
		status       string
	}{
		{"compute", record.ComputeID, record.Name + " compute", record.ComputeProviderID, "running"},
		{"storage", record.StorageID, record.Name + " storage", record.StorageProviderID, "available"},
		{"attachment", record.AttachmentID, record.Name + " attachment", record.AttachProviderID, "attached"},
		{"route", record.WorkspaceID, record.Name + " route", record.RouteProviderID, "ready"},
	}
	for _, view := range views {
		id, err := randomStoreID("mrv")
		if err != nil {
			return err
		}
		metadata, err := json.Marshal(map[string]string{
			"providerResourceId": view.providerID,
			"workspaceUrl":       record.RouteURL,
		})
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO managed_resource_views
			  (id, organization_id, resource_type, resource_id, display_name, provider, status, policy_state, workspace_id, billing_account_id, last_seen_at, metadata)
			VALUES ($1, $2, $3, $4, $5, 'fabric', $6, 'managed', $7, $8, now(), $9)
			ON CONFLICT (organization_id, resource_type, resource_id)
			DO UPDATE SET display_name = EXCLUDED.display_name,
			              provider = EXCLUDED.provider,
			              status = EXCLUDED.status,
			              policy_state = EXCLUDED.policy_state,
			              workspace_id = EXCLUDED.workspace_id,
			              billing_account_id = EXCLUDED.billing_account_id,
			              last_seen_at = EXCLUDED.last_seen_at,
			              metadata = EXCLUDED.metadata,
			              updated_at = now()
		`, id, record.OrganizationID, view.resourceType, view.resourceID, view.displayName, view.status, record.WorkspaceID, record.BillingAccountID, metadata); err != nil {
			return err
		}
	}
	if record.RouteURL == "" {
		return fmt.Errorf("workspace route URL is required")
	}
	return tx.Commit(ctx)
}

func (s *WorkspaceStore) Handoff(ctx context.Context, request workspace.HandoffRequest) (workspace.HandoffResult, error) {
	var result workspace.HandoffResult
	err := s.pool.QueryRow(ctx, `
		SELECT w.id, w.state, COALESCE(v.metadata->>'workspaceUrl', '')
		FROM workspaces w
		JOIN workspace_tokens t ON t.workspace_id = w.id
		LEFT JOIN managed_resource_views v ON v.workspace_id = w.id AND v.resource_type = 'route'
		WHERE w.id = $1
		  AND t.token_hash = $2
		  AND t.status = 'active'
		  AND w.state IN ('running', 'configured')
		LIMIT 1
	`, request.WorkspaceID, auth.HashToken(request.Token)).Scan(&result.WorkspaceID, &result.State, &result.URL)
	if err != nil {
		return workspace.HandoffResult{}, err
	}
	if result.URL == "" {
		return workspace.HandoffResult{}, fmt.Errorf("workspace route URL is missing")
	}
	return result, nil
}

func (s *WorkspaceStore) RuntimeForAction(ctx context.Context, request workspace.ActionRequest) (workspace.RuntimeRecord, error) {
	var record workspace.RuntimeRecord
	err := s.pool.QueryRow(ctx, `
		SELECT w.id, w.billing_account_id, COALESCE(w.compute_id, ''), COALESCE(w.storage_id, ''), w.state
		FROM workspaces w
		JOIN organizations o ON o.billing_account_id = w.billing_account_id
		JOIN memberships m ON m.organization_id = o.id
		WHERE w.id = $1
		  AND m.user_id = $2
		  AND m.status = 'active'
		LIMIT 1
	`, request.WorkspaceID, request.ActorUserID).Scan(
		&record.WorkspaceID,
		&record.BillingAccountID,
		&record.ComputeID,
		&record.StorageID,
		&record.State,
	)
	if err != nil {
		return workspace.RuntimeRecord{}, err
	}
	record.ActorUserID = request.ActorUserID
	return record, nil
}

func (s *WorkspaceStore) UpdateWorkspaceState(ctx context.Context, change workspace.StateChange) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `
		UPDATE workspaces
		SET state = $1, updated_at = now()
		WHERE id = $2
	`, change.State, change.WorkspaceID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE managed_resource_views
		SET status = $1, updated_at = now()
		WHERE workspace_id = $2 AND resource_type = 'route'
	`, change.State, change.WorkspaceID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *WorkspaceStore) ReplaceActiveToken(ctx context.Context, change workspace.TokenChange) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `
		UPDATE workspace_tokens
		SET status = 'deleted', updated_at = now()
		WHERE workspace_id = $1 AND status = 'active'
	`, change.WorkspaceID); err != nil {
		return err
	}
	id, err := randomStoreID("token")
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO workspace_tokens (id, workspace_id, token_hash, status)
		VALUES ($1, $2, $3, 'active')
	`, id, change.WorkspaceID, auth.HashToken(change.Token)); err != nil {
		return err
	}
	if change.URL != "" {
		metadata, err := json.Marshal(map[string]string{"workspaceUrl": change.URL})
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE managed_resource_views
			SET metadata = metadata || $1::jsonb, updated_at = now()
			WHERE workspace_id = $2 AND resource_type = 'route'
		`, metadata, change.WorkspaceID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *WorkspaceStore) DeleteActiveToken(ctx context.Context, request workspace.ActionRequest) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE workspace_tokens
		SET status = 'deleted', updated_at = now()
		WHERE workspace_id = $1 AND status = 'active'
	`, request.WorkspaceID)
	return err
}

var _ workspace.Repository = (*WorkspaceStore)(nil)
