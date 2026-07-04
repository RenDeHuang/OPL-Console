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
	AdminOrganizations(ctx context.Context) ([]OrganizationView, error)
	AdminTeams(ctx context.Context) ([]TeamView, error)
	AdminRoles(ctx context.Context) ([]RoleView, error)
	AdminManagedResources(ctx context.Context) ([]ManagedResourceView, error)
	WalletForUser(ctx context.Context, userID string) (WalletView, error)
	BillingLedgerForUser(ctx context.Context, userID string) ([]BillingLedgerEntryView, error)
	SupportTicketsForUser(ctx context.Context, userID string) ([]SupportTicketView, error)
	CreateSupportTicket(ctx context.Context, userID string, request CreateSupportTicketRequest) (SupportTicketView, error)
	AdminPolicies(ctx context.Context) ([]PolicyView, error)
	CreatePolicy(ctx context.Context, actorUserID string, request CreatePolicyRequest) (PolicyView, error)
	AdminApprovals(ctx context.Context) ([]ApprovalView, error)
	DecideApproval(ctx context.Context, actorUserID string, request ApprovalDecisionRequest) (ApprovalView, error)
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

func (s *Service) AdminOrganizations(ctx context.Context) ([]OrganizationView, error) {
	return s.repo.AdminOrganizations(ctx)
}

func (s *Service) AdminTeams(ctx context.Context) ([]TeamView, error) {
	return s.repo.AdminTeams(ctx)
}

func (s *Service) AdminRoles(ctx context.Context) ([]RoleView, error) {
	return s.repo.AdminRoles(ctx)
}

func (s *Service) AdminManagedResources(ctx context.Context) ([]ManagedResourceView, error) {
	return s.repo.AdminManagedResources(ctx)
}

func (s *Service) Wallet(ctx context.Context, user auth.User) (WalletView, error) {
	return s.repo.WalletForUser(ctx, user.ID)
}

func (s *Service) BillingLedger(ctx context.Context, user auth.User) ([]BillingLedgerEntryView, error) {
	return s.repo.BillingLedgerForUser(ctx, user.ID)
}

func (s *Service) SupportTickets(ctx context.Context, user auth.User) ([]SupportTicketView, error) {
	return s.repo.SupportTicketsForUser(ctx, user.ID)
}

func (s *Service) CreateSupportTicket(ctx context.Context, user auth.User, request CreateSupportTicketRequest) (SupportTicketView, error) {
	return s.repo.CreateSupportTicket(ctx, user.ID, request)
}

func (s *Service) AdminPolicies(ctx context.Context) ([]PolicyView, error) {
	return s.repo.AdminPolicies(ctx)
}

func (s *Service) CreatePolicy(ctx context.Context, user auth.User, request CreatePolicyRequest) (PolicyView, error) {
	return s.repo.CreatePolicy(ctx, user.ID, request)
}

func (s *Service) AdminApprovals(ctx context.Context) ([]ApprovalView, error) {
	return s.repo.AdminApprovals(ctx)
}

func (s *Service) DecideApproval(ctx context.Context, user auth.User, request ApprovalDecisionRequest) (ApprovalView, error) {
	return s.repo.DecideApproval(ctx, user.ID, request)
}
