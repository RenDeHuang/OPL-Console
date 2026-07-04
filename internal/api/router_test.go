package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouteManifestMatchesMedoplConsole(t *testing.T) {
	want := []string{
		"GET /api/healthz",
		"POST /api/auth/login",
		"POST /api/auth/operator-login",
		"POST /api/auth/logout",
		"GET /api/auth/me",
		"GET /api/state",
		"GET /api/operator/summary",
		"POST /api/operator/cleanup-workspace-access",
		"GET /api/management/state",
		"POST /api/billing/topups",
		"POST /api/organizations",
		"POST /api/users",
		"POST /api/users/disable",
		"POST /api/users/delete",
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
	if len(RouteManifest) != len(want) {
		t.Fatalf("manifest length = %d, want %d: %#v", len(RouteManifest), len(want), RouteManifest)
	}
	for i := range want {
		if RouteManifest[i] != want[i] {
			t.Fatalf("manifest[%d] = %q, want %q", i, RouteManifest[i], want[i])
		}
	}
}

func TestLegacyCompatibilityRoutesAreNotRegistered(t *testing.T) {
	handler := NewRouter(Dependencies{})
	legacyRoutes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/me"},
		{http.MethodGet, "/api/packages"},
		{http.MethodGet, "/api/workspaces"},
		{http.MethodGet, "/api/workspaces/ws-alpha"},
		{http.MethodPost, "/api/workspaces/ws-alpha/tokens/reset"},
		{http.MethodGet, "/api/billing/wallet"},
		{http.MethodGet, "/api/admin/users"},
		{http.MethodGet, "/api/admin/ledger"},
		{http.MethodGet, "/workspaces"},
		{http.MethodGet, "/billing"},
	}

	for _, route := range legacyRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			request := httptest.NewRequest(route.method, route.path, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code < 400 {
				t.Fatalf("status = %d, want legacy route rejected", response.Code)
			}
		})
	}
}

func TestHealthz(t *testing.T) {
	handler := NewRouter(Dependencies{})
	request := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	if body := response.Body.String(); body != "{\"ok\":true}\n" {
		t.Fatalf("body = %q", body)
	}
}

func TestOwnerLoginReturnsLabOwnerSession(t *testing.T) {
	handler := NewRouter(Dependencies{})
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"email":"owner@opl.local","password":"password"}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	var payload struct {
		User struct {
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"user"`
		CSRFToken string `json:"csrfToken"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if payload.User.Email != "owner@opl.local" || payload.User.Role != "lab_owner" {
		t.Fatalf("user = %+v, want owner lab_owner", payload.User)
	}
	if payload.CSRFToken == "" {
		t.Fatalf("csrf token missing")
	}
}

func TestAdminLoginReturnsAdminSession(t *testing.T) {
	handler := NewRouter(Dependencies{})
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"email":"admin@opl.local","password":"password"}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	var payload struct {
		User struct {
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"user"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if payload.User.Email != "admin@opl.local" || payload.User.Role != "admin" {
		t.Fatalf("user = %+v, want admin", payload.User)
	}
}

func TestAdminOperatorLoginReturnsAdminSession(t *testing.T) {
	handler := NewRouter(Dependencies{})
	request := httptest.NewRequest(http.MethodPost, "/api/auth/operator-login", bytes.NewBufferString(`{"email":"admin@opl.local","password":"password","operatorToken":"operator-dev-token"}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	var payload struct {
		User struct {
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"user"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode operator login: %v", err)
	}
	if payload.User.Email != "admin@opl.local" || payload.User.Role != "admin" {
		t.Fatalf("user = %+v, want admin", payload.User)
	}
}

func TestLoginRejectsWrongCredentials(t *testing.T) {
	handler := NewRouter(Dependencies{})
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"email":"owner@opl.local","password":"wrong"}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func TestRuntimeReadiness(t *testing.T) {
	handler := NewRouter(Dependencies{RuntimeReady: func() Readiness {
		return Readiness{Ready: true, Checks: map[string]bool{"postgres": true}}
	}})
	request := httptest.NewRequest(http.MethodGet, "/api/runtime/readiness", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	readiness := decodeReadiness(t, response)
	if !readiness.Ready {
		t.Fatalf("ready = %v", readiness.Ready)
	}
	if !readiness.Checks["postgres"] {
		t.Fatalf("postgres = %v", readiness.Checks["postgres"])
	}
}

func TestProductionReadiness(t *testing.T) {
	handler := NewRouter(Dependencies{ProductionReady: func() Readiness {
		return Readiness{Ready: true, Checks: map[string]bool{"production_config": true}}
	}})
	request := httptest.NewRequest(http.MethodGet, "/api/production/readiness", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	readiness := decodeReadiness(t, response)
	if !readiness.Ready {
		t.Fatalf("ready = %v", readiness.Ready)
	}
	if !readiness.Checks["production_config"] {
		t.Fatalf("production_config = %v", readiness.Checks["production_config"])
	}
}

func TestRuntimeReadinessFallback(t *testing.T) {
	handler := NewRouter(Dependencies{})
	request := httptest.NewRequest(http.MethodGet, "/api/runtime/readiness", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	readiness := decodeReadiness(t, response)
	if readiness.Ready {
		t.Fatalf("ready = %v", readiness.Ready)
	}
	configured, ok := readiness.Checks["configured"]
	if !ok {
		t.Fatalf("configured check missing")
	}
	if configured {
		t.Fatalf("configured = %v", configured)
	}
}

func decodeReadiness(t *testing.T, response *httptest.ResponseRecorder) Readiness {
	t.Helper()

	var readiness Readiness
	if err := json.NewDecoder(response.Body).Decode(&readiness); err != nil {
		t.Fatalf("decode readiness: %v", err)
	}
	return readiness
}
