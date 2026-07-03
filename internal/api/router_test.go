package api

import (
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
}
