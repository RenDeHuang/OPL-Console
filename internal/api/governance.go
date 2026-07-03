package api

import (
	"context"
	"net/http"

	"github.com/RenDeHuang/opl-console/internal/auth"
	"github.com/RenDeHuang/opl-console/internal/console"
)

type GovernanceService interface {
	Me(ctx context.Context, user auth.User) (console.Me, error)
	Packages(ctx context.Context) ([]console.Package, error)
	Workspaces(ctx context.Context, user auth.User) ([]console.ManagedWorkspace, error)
	AdminUsers(ctx context.Context) ([]console.UserView, error)
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
