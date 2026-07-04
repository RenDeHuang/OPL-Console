package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

var RouteManifest = []string{
	"GET /api/healthz",
	"POST /api/auth/login",
	"POST /api/auth/operator-login",
	"POST /api/auth/logout",
	"GET /api/auth/me",
	"GET /api/state",
	"GET /api/operator/summary",
	"GET /api/management/state",
	"POST /api/billing/topups",
	"POST /api/organizations",
	"POST /api/users",
	"POST /api/organizations/members",
	"GET /api/compute-pools",
	"GET /api/compute-allocations",
	"GET /api/compute-allocations/:id",
	"POST /api/compute-allocations",
	"POST /api/compute-allocations/:id/destroy",
	"POST /api/storage-volumes",
	"POST /api/storage-volumes/destroy",
	"POST /api/storage-attachments",
	"POST /api/storage-attachments/detach",
	"POST /api/workspaces",
	"POST /api/workspaces/reset-token",
	"POST /api/workspaces/delete-token",
	"POST /api/billing/request-usage",
	"POST /api/billing/reconciliation",
	"GET /api/ledger/task-receipts",
	"POST /api/ledger/task-receipts",
	"GET /api/runtime/readiness",
	"GET /api/production/readiness",
	"POST /api/workspaces/runtime-status",
	"GET /api/support/tickets",
	"POST /api/support/tickets",
}

type Dependencies struct {
	RuntimeReady    func() Readiness
	ProductionReady func() Readiness
}

func NewRouter(deps Dependencies) http.Handler {
	router := chi.NewRouter()
	router.Get("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
	router.Post("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		body := readBody(w, r)
		email, _ := body["email"].(string)
		password, _ := body["password"].(string)
		if password != "password" {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid_credentials"})
			return
		}
		if email == "admin@opl.local" {
			writeJSON(w, http.StatusOK, map[string]any{
				"user": map[string]any{
					"id":        "user-demo-admin",
					"email":     "admin@opl.local",
					"name":      "OPL 管理员",
					"role":      "admin",
					"accountId": "acct-operator",
				},
				"csrfToken": randomToken(),
			})
			return
		}
		if email != "owner@opl.local" {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid_credentials"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"user": map[string]any{
				"id":        "user-demo-owner",
				"email":     email,
				"name":      "OPL 所有者",
				"role":      "lab_owner",
				"accountId": "acct-demo",
			},
			"csrfToken": randomToken(),
		})
	})
	router.Post("/api/auth/operator-login", func(w http.ResponseWriter, r *http.Request) {
		body := readBody(w, r)
		email, _ := body["email"].(string)
		password, _ := body["password"].(string)
		operatorToken, _ := body["operatorToken"].(string)
		if email != "admin@opl.local" || password != "password" || operatorToken != "operator-dev-token" {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid_operator_credentials"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"user": map[string]any{
				"id":        "user-demo-admin",
				"email":     "admin@opl.local",
				"name":      "OPL 管理员",
				"role":      "admin",
				"accountId": "acct-operator",
			},
			"csrfToken": randomToken(),
		})
	})
	router.Post("/api/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
	router.Get("/api/auth/me", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"user": map[string]any{
				"id":        "user-demo-owner",
				"email":     "owner@opl.local",
				"name":      "OPL 所有者",
				"role":      "lab_owner",
				"accountId": "acct-demo",
			},
			"csrfToken": randomToken(),
		})
	})
	router.Get("/api/state", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, demoConsoleState())
	})
	router.Get("/api/operator/summary", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, demoOperatorSummary())
	})
	router.Get("/api/management/state", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, demoManagementState())
	})
	router.Post("/api/billing/topups", acceptedAction("billing_topup_recorded"))
	router.Post("/api/organizations", acceptedAction("organization_created"))
	router.Post("/api/users", acceptedAction("user_created"))
	router.Post("/api/organizations/members", acceptedAction("organization_member_added"))
	router.Get("/api/compute-pools", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"computePools": demoComputePools()})
	})
	router.Get("/api/compute-allocations", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"computeAllocations": demoComputeAllocations()})
	})
	router.Get("/api/compute-allocations/{id}", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"computeAllocation": map[string]any{
			"id":        chi.URLParam(r, "id"),
			"name":      "独立计算",
			"status":    "running",
			"accountId": "acct-demo",
		}})
	})
	router.Post("/api/compute-allocations", acceptedAction("compute_allocation_requested"))
	router.Post("/api/compute-allocations/{id}/destroy", acceptedAction("compute_allocation_destroy_requested"))
	router.Post("/api/storage-volumes", acceptedAction("storage_volume_requested"))
	router.Post("/api/storage-volumes/destroy", acceptedAction("storage_volume_destroy_requested"))
	router.Post("/api/storage-attachments", acceptedAction("storage_attachment_requested"))
	router.Post("/api/storage-attachments/detach", acceptedAction("storage_attachment_detach_requested"))
	router.Post("/api/workspaces", acceptedAction("workspace_create_requested"))
	router.Post("/api/workspaces/reset-token", acceptedAction("workspace_token_reset"))
	router.Post("/api/workspaces/delete-token", acceptedAction("workspace_token_deleted"))
	router.Post("/api/billing/request-usage", acceptedAction("request_usage_recorded"))
	router.Post("/api/billing/reconciliation", acceptedAction("billing_reconciliation_recorded"))
	router.Get("/api/ledger/task-receipts", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"receipts": []map[string]any{}})
	})
	router.Post("/api/ledger/task-receipts", acceptedAction("task_receipt_recorded"))
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
	router.Post("/api/workspaces/runtime-status", acceptedAction("workspace_runtime_status_recorded"))
	router.Get("/api/support/tickets", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"tickets": []map[string]any{
			{
				"id":          "ticket-demo-1",
				"title":       "工作区开通问题",
				"category":    "workspace",
				"status":      "open",
				"createdAt":   time.Now().UTC().Format(time.RFC3339),
				"workspaceId": "ws-alpha",
			},
		}})
	})
	router.Post("/api/support/tickets", acceptedAction("support_ticket_created"))
	return router
}

func acceptedAction(action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusAccepted, map[string]any{
			"ok":      true,
			"action":  action,
			"request": readBody(w, r),
		})
	}
}

func readBody(w http.ResponseWriter, r *http.Request) map[string]any {
	defer r.Body.Close()
	var body map[string]any
	if r.Body == nil || r.ContentLength == 0 {
		return map[string]any{}
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
		return map[string]any{}
	}
	return body
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func randomToken() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "csrf-demo-token"
	}
	return hex.EncodeToString(bytes[:])
}
