package console

import (
	"context"
	"testing"

	"github.com/RenDeHuang/opl-console/internal/auth"
)

type fakeRepository struct {
	user         UserView
	organization OrganizationView
	packages     []Package
	workspaces   []ManagedWorkspace
	adminUsers   []UserView
	orgs         []OrganizationView
	teams        []TeamView
	roles        []RoleView
	resources    []ManagedResourceView
	wallet       WalletView
	ledger       []BillingLedgerEntryView
	tickets      []SupportTicketView
	policies     []PolicyView
	approvals    []ApprovalView

	createdTicket CreateSupportTicketRequest
	createdPolicy CreatePolicyRequest
	decision      ApprovalDecisionRequest
}

func (f fakeRepository) UserByID(ctx context.Context, userID string) (UserView, error) {
	return f.user, nil
}

func (f fakeRepository) PrimaryOrganizationForUser(ctx context.Context, userID string) (OrganizationView, error) {
	return f.organization, nil
}

func (f fakeRepository) Packages(ctx context.Context) ([]Package, error) {
	return f.packages, nil
}

func (f fakeRepository) WorkspacesForUser(ctx context.Context, userID string) ([]ManagedWorkspace, error) {
	return f.workspaces, nil
}

func (f fakeRepository) AdminUsers(ctx context.Context) ([]UserView, error) {
	return f.adminUsers, nil
}

func (f fakeRepository) AdminOrganizations(ctx context.Context) ([]OrganizationView, error) {
	return f.orgs, nil
}

func (f fakeRepository) AdminTeams(ctx context.Context) ([]TeamView, error) {
	return f.teams, nil
}

func (f fakeRepository) AdminRoles(ctx context.Context) ([]RoleView, error) {
	return f.roles, nil
}

func (f fakeRepository) AdminManagedResources(ctx context.Context) ([]ManagedResourceView, error) {
	return f.resources, nil
}

func (f *fakeRepository) WalletForUser(ctx context.Context, userID string) (WalletView, error) {
	return f.wallet, nil
}

func (f *fakeRepository) BillingLedgerForUser(ctx context.Context, userID string) ([]BillingLedgerEntryView, error) {
	return f.ledger, nil
}

func (f *fakeRepository) SupportTicketsForUser(ctx context.Context, userID string) ([]SupportTicketView, error) {
	return f.tickets, nil
}

func (f *fakeRepository) CreateSupportTicket(ctx context.Context, userID string, request CreateSupportTicketRequest) (SupportTicketView, error) {
	f.createdTicket = request
	return SupportTicketView{ID: "ticket-created", Subject: request.Subject, Status: "open"}, nil
}

func (f *fakeRepository) AdminPolicies(ctx context.Context) ([]PolicyView, error) {
	return f.policies, nil
}

func (f *fakeRepository) CreatePolicy(ctx context.Context, actorUserID string, request CreatePolicyRequest) (PolicyView, error) {
	f.createdPolicy = request
	return PolicyView{ID: "policy-created", Name: request.Name, Status: "active"}, nil
}

func (f *fakeRepository) AdminApprovals(ctx context.Context) ([]ApprovalView, error) {
	return f.approvals, nil
}

func (f *fakeRepository) DecideApproval(ctx context.Context, actorUserID string, request ApprovalDecisionRequest) (ApprovalView, error) {
	f.decision = request
	return ApprovalView{ID: request.ApprovalID, Status: request.Decision}, nil
}

func TestServiceMeCombinesUserAndOrganization(t *testing.T) {
	service := NewService(&fakeRepository{
		user:         UserView{ID: "usr-owner", Email: "owner@opl.local", Role: "owner", Status: "active"},
		organization: OrganizationView{ID: "org-alpha", Name: "Alpha Lab", Status: "active"},
	})

	me, err := service.Me(context.Background(), auth.User{ID: "usr-owner"})
	if err != nil {
		t.Fatalf("Me: %v", err)
	}

	if me.User.ID != "usr-owner" || me.Organization.ID != "org-alpha" {
		t.Fatalf("me = %#v", me)
	}
}

func TestServiceExposesReadModelLists(t *testing.T) {
	service := NewService(&fakeRepository{
		packages:   []Package{{ID: "basic", Name: "Basic Workspace"}},
		workspaces: []ManagedWorkspace{{ID: "ws-alpha", Name: "Alpha Workspace"}},
		adminUsers: []UserView{{ID: "usr-admin", Email: "admin@opl.local"}},
	})

	packages, err := service.Packages(context.Background())
	if err != nil {
		t.Fatalf("Packages: %v", err)
	}
	workspaces, err := service.Workspaces(context.Background(), auth.User{ID: "usr-owner"})
	if err != nil {
		t.Fatalf("Workspaces: %v", err)
	}
	users, err := service.AdminUsers(context.Background())
	if err != nil {
		t.Fatalf("AdminUsers: %v", err)
	}

	if len(packages) != 1 || len(workspaces) != 1 || len(users) != 1 {
		t.Fatalf("packages=%#v workspaces=%#v users=%#v", packages, workspaces, users)
	}
}

func TestServiceExposesGovernanceFacades(t *testing.T) {
	repo := &fakeRepository{
		wallet:    WalletView{BillingAccountID: "billing-alpha", BalanceFen: 1000},
		ledger:    []BillingLedgerEntryView{{ID: "ledger-alpha", Kind: "compute_hold"}},
		tickets:   []SupportTicketView{{ID: "ticket-alpha", Subject: "Need help"}},
		policies:  []PolicyView{{ID: "policy-alpha", Name: "Policy"}},
		approvals: []ApprovalView{{ID: "approval-alpha", Status: "pending"}},
	}
	service := NewService(repo)
	user := auth.User{ID: "usr-owner"}

	wallet, err := service.Wallet(context.Background(), user)
	if err != nil {
		t.Fatalf("Wallet: %v", err)
	}
	ledger, err := service.BillingLedger(context.Background(), user)
	if err != nil {
		t.Fatalf("BillingLedger: %v", err)
	}
	tickets, err := service.SupportTickets(context.Background(), user)
	if err != nil {
		t.Fatalf("SupportTickets: %v", err)
	}
	createdTicket, err := service.CreateSupportTicket(context.Background(), user, CreateSupportTicketRequest{Subject: "Need help"})
	if err != nil {
		t.Fatalf("CreateSupportTicket: %v", err)
	}
	policies, err := service.AdminPolicies(context.Background())
	if err != nil {
		t.Fatalf("AdminPolicies: %v", err)
	}
	createdPolicy, err := service.CreatePolicy(context.Background(), auth.User{ID: "usr-admin"}, CreatePolicyRequest{Name: "Policy", PolicyType: "workspace_lifecycle"})
	if err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}
	approvals, err := service.AdminApprovals(context.Background())
	if err != nil {
		t.Fatalf("AdminApprovals: %v", err)
	}
	decision, err := service.DecideApproval(context.Background(), auth.User{ID: "usr-admin"}, ApprovalDecisionRequest{ApprovalID: "approval-alpha", Decision: "approved"})
	if err != nil {
		t.Fatalf("DecideApproval: %v", err)
	}

	if wallet.BillingAccountID == "" || len(ledger) != 1 || len(tickets) != 1 || createdTicket.ID == "" ||
		len(policies) != 1 || createdPolicy.ID == "" || len(approvals) != 1 || decision.Status != "approved" {
		t.Fatalf("wallet=%#v ledger=%#v tickets=%#v createdTicket=%#v policies=%#v createdPolicy=%#v approvals=%#v decision=%#v",
			wallet, ledger, tickets, createdTicket, policies, createdPolicy, approvals, decision)
	}
	if repo.createdPolicy.PolicyType != "workspace_lifecycle" || repo.decision.ApprovalID != "approval-alpha" {
		t.Fatalf("repo = %#v", repo)
	}
}
