package store

import (
	"context"

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
