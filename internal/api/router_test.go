package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
