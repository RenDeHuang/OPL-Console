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

	request := httptest.NewRequest(http.MethodGet, "/api/state", nil)
	request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var payload map[string]json.RawMessage
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode state: %v", err)
	}
	for _, key := range []string{"me", "packages", "workspaces", "wallet", "ledger", "tickets"} {
		if len(payload[key]) == 0 {
			t.Fatalf("state missing %s: %#v", key, payload)
		}
	}
}

func TestOwnerGovernanceRoutesRejectMissingSession(t *testing.T) {
	handler := NewRouter(Dependencies{
		Auth:              &fakeAuthService{},
		SessionCookieName: "opl_session",
		Governance:        &fakeGovernanceService{},
	})
	request := httptest.NewRequest(http.MethodGet, "/api/state", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestManagementStateRequiresAdmin(t *testing.T) {
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
			request := httptest.NewRequest(http.MethodGet, "/api/management/state", nil)
			request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d, body = %s", response.Code, tt.want, response.Body.String())
			}
			if response.Code == http.StatusOK {
				var payload map[string]json.RawMessage
				if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
					t.Fatalf("decode management state: %v", err)
				}
				if len(payload["users"]) == 0 {
					t.Fatalf("users missing: %#v", payload)
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

	requestState := httptest.NewRequest(http.MethodGet, "/api/state", nil)
	requestState.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
	stateResponse := httptest.NewRecorder()
	handler.ServeHTTP(stateResponse, requestState)
	if stateResponse.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", stateResponse.Code, stateResponse.Body.String())
	}

	requestTickets := httptest.NewRequest(http.MethodGet, "/api/support/tickets", nil)
	requestTickets.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
	ticketsResponse := httptest.NewRecorder()
	handler.ServeHTTP(ticketsResponse, requestTickets)
	if ticketsResponse.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", ticketsResponse.Code, ticketsResponse.Body.String())
	}
	var ticketsPayload map[string]json.RawMessage
	if err := json.NewDecoder(ticketsResponse.Body).Decode(&ticketsPayload); err != nil {
		t.Fatalf("decode tickets: %v", err)
	}
	if len(ticketsPayload["tickets"]) == 0 {
		t.Fatalf("tickets missing: %#v", ticketsPayload)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/support/tickets", strings.NewReader(`{"subject":"Need help","body":"Workspace is blocked","workspaceId":"ws-alpha"}`))
	request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
	request.Header.Set("x-opl-csrf-token", "csrf-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if governance.createdTicket.Subject != "Need help" || governance.createdTicket.WorkspaceID != "ws-alpha" {
		t.Fatalf("created ticket = %#v", governance.createdTicket)
	}
}

func TestOwnerMutationRejectsMissingCSRF(t *testing.T) {
	handler := NewRouter(Dependencies{
		Auth: &fakeAuthService{session: auth.Session{
			Token:     "session-token",
			CSRFToken: "csrf-token",
			ExpiresAt: time.Now().Add(time.Hour),
			User:      auth.User{ID: "usr-owner", Email: "owner@opl.local", Role: auth.RoleOwner, Status: auth.StatusActive},
		}},
		SessionCookieName: "opl_session",
		Governance:        &fakeGovernanceService{},
	})
	request := httptest.NewRequest(http.MethodPost, "/api/support/tickets", strings.NewReader(`{"subject":"Need help","body":"Workspace is blocked"}`))
	request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestManagementStateIncludesPoliciesAndApprovals(t *testing.T) {
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

	request := httptest.NewRequest(http.MethodGet, "/api/management/state", nil)
	request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var payload map[string]json.RawMessage
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode management state: %v", err)
	}
	if len(payload["policies"]) == 0 || len(payload["approvals"]) == 0 {
		t.Fatalf("management state = %#v", payload)
	}
}

func TestManagementStateIncludesGovernanceReadModels(t *testing.T) {
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

	request := httptest.NewRequest(http.MethodGet, "/api/management/state", nil)
	request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var payload map[string]json.RawMessage
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode management state: %v", err)
	}
	for _, key := range []string{"organizations", "teams", "roles", "resources"} {
		if len(payload[key]) == 0 {
			t.Fatalf("management state missing %s: %#v", key, payload)
		}
	}
}

func TestOperatorSummaryAllowsConfiguredToken(t *testing.T) {
	handler := NewRouter(Dependencies{
		OperatorSummaryToken: "operator-secret",
		Governance: &fakeGovernanceService{
			adminUsers: []console.UserView{{ID: "usr-admin", Email: "admin@opl.local", Role: "admin", Status: "active"}},
			resources:  []console.ManagedResourceView{{ID: "mrv-alpha", OrganizationID: "org-alpha", ResourceType: "compute", ResourceID: "cmp-alpha", Status: "running"}},
		},
	})
	request := httptest.NewRequest(http.MethodGet, "/api/operator/summary", nil)
	request.Header.Set("x-opl-operator-token", "operator-secret")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestOldGovernanceCompatibilityRoutesAreNotMounted(t *testing.T) {
	handler := NewRouter(Dependencies{})
	for _, item := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/me"},
		{http.MethodGet, "/api/packages"},
		{http.MethodGet, "/api/workspaces"},
		{http.MethodGet, "/api/billing/wallet"},
		{http.MethodGet, "/api/billing/ledger"},
		{http.MethodGet, "/api/admin/users"},
		{http.MethodGet, "/api/admin/policies"},
		{http.MethodPost, "/api/admin/approvals/approval-alpha/approve"},
	} {
		t.Run(item.method+" "+item.path, func(t *testing.T) {
			request := httptest.NewRequest(item.method, item.path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusNotFound && response.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}
}
