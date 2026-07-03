package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/RenDeHuang/opl-console/internal/auth"
	"github.com/RenDeHuang/opl-console/internal/console"
)

type fakeGovernanceService struct {
	me         console.Me
	packages   []console.Package
	workspaces []console.ManagedWorkspace
	adminUsers []console.UserView
	orgs       []console.OrganizationView
	teams      []console.TeamView
	roles      []console.RoleView
	resources  []console.ManagedResourceView
	wallet     console.WalletView
	ledger     []console.BillingLedgerEntryView
	tickets    []console.SupportTicketView
	policies   []console.PolicyView
	approvals  []console.ApprovalView

	createdTicket console.CreateSupportTicketRequest
	createdPolicy console.CreatePolicyRequest
	decision      console.ApprovalDecisionRequest
}

func (f fakeGovernanceService) Me(ctx context.Context, user auth.User) (console.Me, error) {
	return f.me, nil
}

func (f fakeGovernanceService) Packages(ctx context.Context) ([]console.Package, error) {
	return f.packages, nil
}

func (f fakeGovernanceService) Workspaces(ctx context.Context, user auth.User) ([]console.ManagedWorkspace, error) {
	return f.workspaces, nil
}

func (f fakeGovernanceService) AdminUsers(ctx context.Context) ([]console.UserView, error) {
	return f.adminUsers, nil
}

func (f fakeGovernanceService) AdminOrganizations(ctx context.Context) ([]console.OrganizationView, error) {
	return f.orgs, nil
}

func (f fakeGovernanceService) AdminTeams(ctx context.Context) ([]console.TeamView, error) {
	return f.teams, nil
}

func (f fakeGovernanceService) AdminRoles(ctx context.Context) ([]console.RoleView, error) {
	return f.roles, nil
}

func (f fakeGovernanceService) AdminManagedResources(ctx context.Context) ([]console.ManagedResourceView, error) {
	return f.resources, nil
}

func (f *fakeGovernanceService) Wallet(ctx context.Context, user auth.User) (console.WalletView, error) {
	return f.wallet, nil
}

func (f *fakeGovernanceService) BillingLedger(ctx context.Context, user auth.User) ([]console.BillingLedgerEntryView, error) {
	return f.ledger, nil
}

func (f *fakeGovernanceService) SupportTickets(ctx context.Context, user auth.User) ([]console.SupportTicketView, error) {
	return f.tickets, nil
}

func (f *fakeGovernanceService) CreateSupportTicket(ctx context.Context, user auth.User, request console.CreateSupportTicketRequest) (console.SupportTicketView, error) {
	f.createdTicket = request
	return console.SupportTicketView{ID: "ticket-created", Subject: request.Subject, Status: "open"}, nil
}

func (f *fakeGovernanceService) AdminPolicies(ctx context.Context) ([]console.PolicyView, error) {
	return f.policies, nil
}

func (f *fakeGovernanceService) CreatePolicy(ctx context.Context, user auth.User, request console.CreatePolicyRequest) (console.PolicyView, error) {
	f.createdPolicy = request
	return console.PolicyView{ID: "policy-created", Name: request.Name, PolicyType: request.PolicyType, Status: "active"}, nil
}

func (f *fakeGovernanceService) AdminApprovals(ctx context.Context) ([]console.ApprovalView, error) {
	return f.approvals, nil
}

func (f *fakeGovernanceService) DecideApproval(ctx context.Context, user auth.User, request console.ApprovalDecisionRequest) (console.ApprovalView, error) {
	f.decision = request
	return console.ApprovalView{ID: request.ApprovalID, Status: request.Decision}, nil
}

func TestOwnerGovernanceRoutesRequireActiveOwnerSession(t *testing.T) {
	authService := &fakeAuthService{session: auth.Session{
		Token:     "session-token",
		CSRFToken: "csrf-token",
		ExpiresAt: time.Now().Add(time.Hour),
		User: auth.User{
			ID:     "usr-owner",
			Email:  "owner@opl.local",
			Role:   auth.RoleOwner,
			Status: auth.StatusActive,
		},
	}}
	handler := NewRouter(Dependencies{
		Auth:              authService,
		SessionCookieName: "opl_session",
		Governance: &fakeGovernanceService{
			me: console.Me{
				User:         console.UserView{ID: "usr-owner", Email: "owner@opl.local", Role: "owner", Status: "active"},
				Organization: console.OrganizationView{ID: "org-alpha", Name: "Alpha Lab", Status: "active"},
			},
			packages: []console.Package{{ID: "basic", Name: "Basic Workspace", CPU: 2, MemoryGB: 4, StorageGB: 10}},
			workspaces: []console.ManagedWorkspace{{
				ID:     "ws-alpha",
				Name:   "Alpha Workspace",
				State:  "running",
				Policy: "managed",
			}},
		},
	})

	for _, path := range []string{"/api/me", "/api/packages", "/api/workspaces"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestOwnerGovernanceRoutesRejectMissingSession(t *testing.T) {
	handler := NewRouter(Dependencies{
		Auth:              &fakeAuthService{},
		SessionCookieName: "opl_session",
		Governance:        &fakeGovernanceService{},
	})
	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestAdminGovernanceRoutesRequireAdmin(t *testing.T) {
	tests := []struct {
		name string
		user auth.User
		want int
	}{
		{
			name: "admin",
			user: auth.User{ID: "usr-admin", Email: "admin@opl.local", Role: auth.RoleAdmin, Status: auth.StatusActive},
			want: http.StatusOK,
		},
		{
			name: "owner",
			user: auth.User{ID: "usr-owner", Email: "owner@opl.local", Role: auth.RoleOwner, Status: auth.StatusActive},
			want: http.StatusForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewRouter(Dependencies{
				Auth: &fakeAuthService{session: auth.Session{
					Token:     "session-token",
					CSRFToken: "csrf-token",
					ExpiresAt: time.Now().Add(time.Hour),
					User:      tt.user,
				}},
				SessionCookieName: "opl_session",
				Governance: &fakeGovernanceService{
					adminUsers: []console.UserView{{ID: "usr-admin", Email: "admin@opl.local", Role: "admin", Status: "active"}},
				},
			})
			request := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
			request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d, body = %s", response.Code, tt.want, response.Body.String())
			}
			if response.Code == http.StatusOK {
				var users []console.UserView
				if err := json.NewDecoder(response.Body).Decode(&users); err != nil {
					t.Fatalf("decode users: %v", err)
				}
				if len(users) != 1 {
					t.Fatalf("users = %#v", users)
				}
			}
		})
	}
}

func TestOwnerBillingAndSupportRoutes(t *testing.T) {
	governance := &fakeGovernanceService{
		wallet:  console.WalletView{BillingAccountID: "billing-alpha", BalanceFen: 1000, FrozenFen: 250, AvailableFen: 750},
		ledger:  []console.BillingLedgerEntryView{{ID: "ledger-alpha", Kind: "compute_hold", AmountFen: -250}},
		tickets: []console.SupportTicketView{{ID: "ticket-alpha", Subject: "Need help", Status: "open"}},
	}
	handler := NewRouter(Dependencies{
		Auth: &fakeAuthService{session: auth.Session{
			Token:     "session-token",
			CSRFToken: "csrf-token",
			ExpiresAt: time.Now().Add(time.Hour),
			User:      auth.User{ID: "usr-owner", Email: "owner@opl.local", Role: auth.RoleOwner, Status: auth.StatusActive},
		}},
		SessionCookieName: "opl_session",
		Governance:        governance,
	})

	for _, path := range []string{"/api/billing/wallet", "/api/billing/ledger", "/api/support/tickets"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}

	request := httptest.NewRequest(http.MethodPost, "/api/support/tickets", strings.NewReader(`{"subject":"Need help","body":"Workspace is blocked","workspaceId":"ws-alpha"}`))
	request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if governance.createdTicket.Subject != "Need help" || governance.createdTicket.WorkspaceID != "ws-alpha" {
		t.Fatalf("created ticket = %#v", governance.createdTicket)
	}
}

func TestAdminPolicyAndApprovalRoutes(t *testing.T) {
	governance := &fakeGovernanceService{
		policies:  []console.PolicyView{{ID: "policy-alpha", Name: "Managed Workspace Approval", Status: "active"}},
		approvals: []console.ApprovalView{{ID: "approval-alpha", Status: "pending"}},
	}
	handler := NewRouter(Dependencies{
		Auth: &fakeAuthService{session: auth.Session{
			Token:     "session-token",
			CSRFToken: "csrf-token",
			ExpiresAt: time.Now().Add(time.Hour),
			User:      auth.User{ID: "usr-admin", Email: "admin@opl.local", Role: auth.RoleAdmin, Status: auth.StatusActive},
		}},
		SessionCookieName: "opl_session",
		Governance:        governance,
	})

	for _, path := range []string{"/api/admin/policies", "/api/admin/approvals"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}

	createPolicy := httptest.NewRequest(http.MethodPost, "/api/admin/policies", strings.NewReader(`{"organizationId":"org-alpha","name":"Managed Workspace Approval","policyType":"workspace_lifecycle","rules":{"requiresApproval":true}}`))
	createPolicy.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
	createPolicyResponse := httptest.NewRecorder()
	handler.ServeHTTP(createPolicyResponse, createPolicy)
	if createPolicyResponse.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", createPolicyResponse.Code, createPolicyResponse.Body.String())
	}
	if governance.createdPolicy.OrganizationID != "org-alpha" || governance.createdPolicy.PolicyType != "workspace_lifecycle" {
		t.Fatalf("created policy = %#v", governance.createdPolicy)
	}

	approve := httptest.NewRequest(http.MethodPost, "/api/admin/approvals/approval-alpha/approve", strings.NewReader(`{"decisionNote":"approved for pilot"}`))
	approve.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
	approveResponse := httptest.NewRecorder()
	handler.ServeHTTP(approveResponse, approve)
	if approveResponse.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", approveResponse.Code, approveResponse.Body.String())
	}
	if governance.decision.ApprovalID != "approval-alpha" || governance.decision.Decision != "approved" {
		t.Fatalf("decision = %#v", governance.decision)
	}
}

func TestAdminGovernanceReadModelRoutes(t *testing.T) {
	handler := NewRouter(Dependencies{
		Auth: &fakeAuthService{session: auth.Session{
			Token:     "session-token",
			CSRFToken: "csrf-token",
			ExpiresAt: time.Now().Add(time.Hour),
			User:      auth.User{ID: "usr-admin", Email: "admin@opl.local", Role: auth.RoleAdmin, Status: auth.StatusActive},
		}},
		SessionCookieName: "opl_session",
		Governance: &fakeGovernanceService{
			orgs:      []console.OrganizationView{{ID: "org-alpha", Name: "Alpha Lab", Status: "active"}},
			teams:     []console.TeamView{{ID: "team-alpha", OrganizationID: "org-alpha", Name: "Platform", Status: "active"}},
			roles:     []console.RoleView{{ID: "role-alpha", OrganizationID: "org-alpha", Name: "Owner", Scope: "organization"}},
			resources: []console.ManagedResourceView{{ID: "mrv-alpha", OrganizationID: "org-alpha", ResourceType: "compute", ResourceID: "cmp-alpha", Status: "running"}},
		},
	})

	for _, path := range []string{"/api/admin/organizations", "/api/admin/teams", "/api/admin/roles", "/api/admin/resources"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}
}
