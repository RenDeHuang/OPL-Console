package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/RenDeHuang/opl-console/internal/auth"
	"github.com/RenDeHuang/opl-console/internal/console"
)

type GovernanceService interface {
	Me(ctx context.Context, user auth.User) (console.Me, error)
	Packages(ctx context.Context) ([]console.Package, error)
	Workspaces(ctx context.Context, user auth.User) ([]console.ManagedWorkspace, error)
	AdminUsers(ctx context.Context) ([]console.UserView, error)
	AdminOrganizations(ctx context.Context) ([]console.OrganizationView, error)
	AdminTeams(ctx context.Context) ([]console.TeamView, error)
	AdminRoles(ctx context.Context) ([]console.RoleView, error)
	AdminManagedResources(ctx context.Context) ([]console.ManagedResourceView, error)
	Wallet(ctx context.Context, user auth.User) (console.WalletView, error)
	BillingLedger(ctx context.Context, user auth.User) ([]console.BillingLedgerEntryView, error)
	SupportTickets(ctx context.Context, user auth.User) ([]console.SupportTicketView, error)
	CreateSupportTicket(ctx context.Context, user auth.User, request console.CreateSupportTicketRequest) (console.SupportTicketView, error)
	AdminPolicies(ctx context.Context) ([]console.PolicyView, error)
	CreatePolicy(ctx context.Context, user auth.User, request console.CreatePolicyRequest) (console.PolicyView, error)
	AdminApprovals(ctx context.Context) ([]console.ApprovalView, error)
	DecideApproval(ctx context.Context, user auth.User, request console.ApprovalDecisionRequest) (console.ApprovalView, error)
}

func mountGovernanceRoutes(router Router, deps Dependencies) {
	router.Get("/api/state", func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireOwner(w, r, deps)
		if !ok {
			return
		}
		result, err := ownerState(r.Context(), governanceService(deps), session.User)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "state_read_failed"})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	router.Get("/api/operator/summary", func(w http.ResponseWriter, r *http.Request) {
		if !authorizedOperator(r, deps) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "operator_summary_token_invalid"})
			return
		}
		service := governanceService(deps)
		users, err := service.AdminUsers(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "operator_summary_failed"})
			return
		}
		resources, err := service.AdminManagedResources(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "operator_summary_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"users":               len(users),
			"managedResources":    len(resources),
			"runtimeReadiness":    readinessOrDefault(deps.RuntimeReady),
			"productionReadiness": readinessOrDefault(deps.ProductionReady),
		})
	})

	router.Get("/api/management/state", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, deps); !ok {
			return
		}
		result, err := managementState(r.Context(), governanceService(deps))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "management_state_failed"})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	router.Post("/api/organizations", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, deps); !ok {
			return
		}
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "organization_create_not_implemented"})
	})

	router.Post("/api/users", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, deps); !ok {
			return
		}
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "user_create_not_implemented"})
	})

	router.Post("/api/organizations/members", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, deps); !ok {
			return
		}
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "organization_member_create_not_implemented"})
	})

	router.Get("/api/support/tickets", func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireOwner(w, r, deps)
		if !ok {
			return
		}
		result, err := governanceService(deps).SupportTickets(r.Context(), session.User)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "read_model_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"tickets": result})
	})

	router.Post("/api/support/tickets", func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireOwner(w, r, deps)
		if !ok {
			return
		}
		var payload console.CreateSupportTicketRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}
		result, err := governanceService(deps).CreateSupportTicket(r.Context(), session.User, payload)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "support_ticket_create_failed"})
			return
		}
		writeJSON(w, http.StatusCreated, result)
	})
}

func ownerState(ctx context.Context, service GovernanceService, user auth.User) (map[string]any, error) {
	me, err := service.Me(ctx, user)
	if err != nil {
		return nil, err
	}
	packages, err := service.Packages(ctx)
	if err != nil {
		return nil, err
	}
	workspaces, err := service.Workspaces(ctx, user)
	if err != nil {
		return nil, err
	}
	wallet, err := service.Wallet(ctx, user)
	if err != nil {
		return nil, err
	}
	ledger, err := service.BillingLedger(ctx, user)
	if err != nil {
		return nil, err
	}
	tickets, err := service.SupportTickets(ctx, user)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"me":         me,
		"packages":   packages,
		"workspaces": workspaces,
		"wallet":     wallet,
		"ledger":     ledger,
		"tickets":    tickets,
	}, nil
}

func managementState(ctx context.Context, service GovernanceService) (map[string]any, error) {
	users, err := service.AdminUsers(ctx)
	if err != nil {
		return nil, err
	}
	organizations, err := service.AdminOrganizations(ctx)
	if err != nil {
		return nil, err
	}
	teams, err := service.AdminTeams(ctx)
	if err != nil {
		return nil, err
	}
	roles, err := service.AdminRoles(ctx)
	if err != nil {
		return nil, err
	}
	resources, err := service.AdminManagedResources(ctx)
	if err != nil {
		return nil, err
	}
	policies, err := service.AdminPolicies(ctx)
	if err != nil {
		return nil, err
	}
	approvals, err := service.AdminApprovals(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"users":         users,
		"organizations": organizations,
		"teams":         teams,
		"roles":         roles,
		"resources":     resources,
		"policies":      policies,
		"approvals":     approvals,
	}, nil
}

func readinessOrDefault(check func() Readiness) Readiness {
	if check == nil {
		return Readiness{Ready: false, Checks: map[string]bool{"configured": false}}
	}
	return check()
}

func authorizedOperator(r *http.Request, deps Dependencies) bool {
	if session, ok := optionalSession(r, deps); ok && auth.CanAccessAdmin(session.User) {
		return true
	}
	if deps.OperatorSummaryToken == "" {
		return false
	}
	token := r.Header.Get("x-opl-operator-token")
	if token == "" {
		token = r.URL.Query().Get("operatorToken")
	}
	return token != "" && token == deps.OperatorSummaryToken
}

func requireOwner(w http.ResponseWriter, r *http.Request, deps Dependencies) (auth.Session, bool) {
	session, ok := sessionFromRequest(w, r, deps)
	if !ok {
		return auth.Session{}, false
	}
	if !requireCSRF(w, r, session) {
		return auth.Session{}, false
	}
	if !auth.CanAccessOwner(session.User) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return auth.Session{}, false
	}
	return session, true
}

func requireAdmin(w http.ResponseWriter, r *http.Request, deps Dependencies) (auth.Session, bool) {
	session, ok := sessionFromRequest(w, r, deps)
	if !ok {
		return auth.Session{}, false
	}
	if !requireCSRF(w, r, session) {
		return auth.Session{}, false
	}
	if !auth.CanAccessAdmin(session.User) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return auth.Session{}, false
	}
	return session, true
}

func requireCSRF(w http.ResponseWriter, r *http.Request, session auth.Session) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return true
	}
	token := r.Header.Get("x-opl-csrf-token")
	if token == "" {
		token = r.Header.Get("x-opl-csrf")
	}
	if token == "" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "csrf_token_required"})
		return false
	}
	if token == session.CSRFToken || auth.HashToken(token) == session.CSRFToken {
		return true
	}
	writeJSON(w, http.StatusForbidden, map[string]string{"error": "csrf_token_invalid"})
	return false
}

func sessionFromRequest(w http.ResponseWriter, r *http.Request, deps Dependencies) (auth.Session, bool) {
	if deps.Auth == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not_authenticated"})
		return auth.Session{}, false
	}
	cookieName := deps.SessionCookieName
	if cookieName == "" {
		cookieName = defaultSessionCookieName
	}
	cookie, err := r.Cookie(cookieName)
	if err != nil || cookie.Value == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not_authenticated"})
		return auth.Session{}, false
	}
	session, err := deps.Auth.Session(r.Context(), cookie.Value)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not_authenticated"})
		return auth.Session{}, false
	}
	return session, true
}

func optionalSession(r *http.Request, deps Dependencies) (auth.Session, bool) {
	if deps.Auth == nil {
		return auth.Session{}, false
	}
	cookieName := deps.SessionCookieName
	if cookieName == "" {
		cookieName = defaultSessionCookieName
	}
	cookie, err := r.Cookie(cookieName)
	if err != nil || cookie.Value == "" {
		return auth.Session{}, false
	}
	session, err := deps.Auth.Session(r.Context(), cookie.Value)
	if err != nil {
		return auth.Session{}, false
	}
	return session, true
}

func governanceService(deps Dependencies) GovernanceService {
	if deps.Governance != nil {
		return deps.Governance
	}
	return emptyGovernanceService{}
}

type emptyGovernanceService struct{}

func (emptyGovernanceService) Me(ctx context.Context, user auth.User) (console.Me, error) {
	return console.Me{
		User: console.UserView{
			ID:     user.ID,
			Email:  user.Email,
			Role:   string(user.Role),
			Status: string(user.Status),
		},
	}, nil
}

func (emptyGovernanceService) Packages(ctx context.Context) ([]console.Package, error) {
	return []console.Package{}, nil
}

func (emptyGovernanceService) Workspaces(ctx context.Context, user auth.User) ([]console.ManagedWorkspace, error) {
	return []console.ManagedWorkspace{}, nil
}

func (emptyGovernanceService) AdminUsers(ctx context.Context) ([]console.UserView, error) {
	return []console.UserView{}, nil
}

func (emptyGovernanceService) AdminOrganizations(ctx context.Context) ([]console.OrganizationView, error) {
	return []console.OrganizationView{}, nil
}

func (emptyGovernanceService) AdminTeams(ctx context.Context) ([]console.TeamView, error) {
	return []console.TeamView{}, nil
}

func (emptyGovernanceService) AdminRoles(ctx context.Context) ([]console.RoleView, error) {
	return []console.RoleView{}, nil
}

func (emptyGovernanceService) AdminManagedResources(ctx context.Context) ([]console.ManagedResourceView, error) {
	return []console.ManagedResourceView{}, nil
}

func (emptyGovernanceService) Wallet(ctx context.Context, user auth.User) (console.WalletView, error) {
	return console.WalletView{}, nil
}

func (emptyGovernanceService) BillingLedger(ctx context.Context, user auth.User) ([]console.BillingLedgerEntryView, error) {
	return []console.BillingLedgerEntryView{}, nil
}

func (emptyGovernanceService) SupportTickets(ctx context.Context, user auth.User) ([]console.SupportTicketView, error) {
	return []console.SupportTicketView{}, nil
}

func (emptyGovernanceService) CreateSupportTicket(ctx context.Context, user auth.User, request console.CreateSupportTicketRequest) (console.SupportTicketView, error) {
	return console.SupportTicketView{}, nil
}

func (emptyGovernanceService) AdminPolicies(ctx context.Context) ([]console.PolicyView, error) {
	return []console.PolicyView{}, nil
}

func (emptyGovernanceService) CreatePolicy(ctx context.Context, user auth.User, request console.CreatePolicyRequest) (console.PolicyView, error) {
	return console.PolicyView{}, nil
}

func (emptyGovernanceService) AdminApprovals(ctx context.Context) ([]console.ApprovalView, error) {
	return []console.ApprovalView{}, nil
}

func (emptyGovernanceService) DecideApproval(ctx context.Context, user auth.User, request console.ApprovalDecisionRequest) (console.ApprovalView, error) {
	return console.ApprovalView{}, nil
}
