package console

import (
	"context"

	"github.com/RenDeHuang/opl-console/internal/auth"
)

type Repository interface {
	UserByID(ctx context.Context, userID string) (UserView, error)
	PrimaryOrganizationForUser(ctx context.Context, userID string) (OrganizationView, error)
	Packages(ctx context.Context) ([]Package, error)
	WorkspacesForUser(ctx context.Context, userID string) ([]ManagedWorkspace, error)
	AdminUsers(ctx context.Context) ([]UserView, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Me(ctx context.Context, user auth.User) (Me, error) {
	userView, err := s.repo.UserByID(ctx, user.ID)
	if err != nil {
		return Me{}, err
	}
	organization, err := s.repo.PrimaryOrganizationForUser(ctx, user.ID)
	if err != nil {
		return Me{}, err
	}
	return Me{User: userView, Organization: organization}, nil
}

func (s *Service) Packages(ctx context.Context) ([]Package, error) {
	return s.repo.Packages(ctx)
}

func (s *Service) Workspaces(ctx context.Context, user auth.User) ([]ManagedWorkspace, error) {
	return s.repo.WorkspacesForUser(ctx, user.ID)
}

func (s *Service) AdminUsers(ctx context.Context) ([]UserView, error) {
	return s.repo.AdminUsers(ctx)
}
