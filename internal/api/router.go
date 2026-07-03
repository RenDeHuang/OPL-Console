package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Router interface {
	Get(pattern string, handlerFn http.HandlerFunc)
	Post(pattern string, handlerFn http.HandlerFunc)
}

type Dependencies struct {
	RuntimeReady      func() Readiness
	ProductionReady   func() Readiness
	Auth              AuthService
	Governance        GovernanceService
	SessionCookieName string
}

func NewRouter(deps Dependencies) http.Handler {
	router := chi.NewRouter()
	router.Get("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
	router.Get("/api/runtime/readiness", func(w http.ResponseWriter, r *http.Request) {
		check := deps.RuntimeReady
		if check == nil {
			check = func() Readiness {
				return Readiness{Ready: false, Checks: map[string]bool{"configured": false}}
			}
		}
		writeJSON(w, http.StatusOK, check())
	})
	router.Get("/api/production/readiness", func(w http.ResponseWriter, r *http.Request) {
		check := deps.ProductionReady
		if check == nil {
			check = func() Readiness {
				return Readiness{Ready: false, Checks: map[string]bool{"configured": false}}
			}
		}
		writeJSON(w, http.StatusOK, check())
	})
	mountAuthRoutes(router, deps)
	mountGovernanceRoutes(router, deps)
	return router
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
