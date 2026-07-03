package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/RenDeHuang/opl-console/internal/auth"
	"github.com/RenDeHuang/opl-console/internal/console"
	"github.com/go-chi/chi/v5"
)

type GovernanceService interface {
	Me(ctx context.Context, user auth.User) (console.Me, error)
	Packages(ctx context.Context) ([]console.Package, error)
	Workspaces(ctx context.Context, user auth.User) ([]console.ManagedWorkspace, error)
	AdminUsers(ctx context.Context) ([]console.UserView, error)
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
	router.Get("/api/me", func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireOwner(w, r, deps)
		if !ok {
			return
		}
		result, err := governanceService(deps).Me(r.Context(), session.User)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "read_model_failed"})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	router.Get("/api/packages", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireOwner(w, r, deps); !ok {
			return
		}
		result, err := governanceService(deps).Packages(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "read_model_failed"})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	router.Get("/api/workspaces", func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireOwner(w, r, deps)
		if !ok {
			return
		}
		result, err := governanceService(deps).Workspaces(r.Context(), session.User)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "read_model_failed"})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	router.Get("/api/admin/users", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, deps); !ok {
			return
		}
		result, err := governanceService(deps).AdminUsers(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "read_model_failed"})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	router.Get("/api/billing/wallet", func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireOwner(w, r, deps)
		if !ok {
			return
		}
		result, err := governanceService(deps).Wallet(r.Context(), session.User)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "read_model_failed"})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	router.Get("/api/billing/ledger", func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireOwner(w, r, deps)
		if !ok {
			return
		}
		result, err := governanceService(deps).BillingLedger(r.Context(), session.User)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "read_model_failed"})
			return
		}
		writeJSON(w, http.StatusOK, result)
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
		writeJSON(w, http.StatusOK, result)
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

	router.Get("/api/admin/policies", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, deps); !ok {
			return
		}
		result, err := governanceService(deps).AdminPolicies(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "read_model_failed"})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	router.Post("/api/admin/policies", func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireAdmin(w, r, deps)
		if !ok {
			return
		}
		var payload console.CreatePolicyRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}
		result, err := governanceService(deps).CreatePolicy(r.Context(), session.User, payload)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "policy_create_failed"})
			return
		}
		writeJSON(w, http.StatusCreated, result)
	})

	router.Get("/api/admin/approvals", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, deps); !ok {
			return
		}
		result, err := governanceService(deps).AdminApprovals(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "read_model_failed"})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	router.Post("/api/admin/approvals/{id}/approve", func(w http.ResponseWriter, r *http.Request) {
		decideApproval(w, r, deps, "approved")
	})
	router.Post("/api/admin/approvals/{id}/reject", func(w http.ResponseWriter, r *http.Request) {
		decideApproval(w, r, deps, "rejected")
	})
}

func decideApproval(w http.ResponseWriter, r *http.Request, deps Dependencies, decision string) {
	session, ok := requireAdmin(w, r, deps)
	if !ok {
		return
	}
	var payload console.ApprovalDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	payload.ApprovalID = chi.URLParam(r, "id")
	payload.Decision = decision
	result, err := governanceService(deps).DecideApproval(r.Context(), session.User, payload)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "approval_decision_failed"})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func requireOwner(w http.ResponseWriter, r *http.Request, deps Dependencies) (auth.Session, bool) {
	session, ok := sessionFromRequest(w, r, deps)
	if !ok {
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
	if !auth.CanAccessAdmin(session.User) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return auth.Session{}, false
	}
	return session, true
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
