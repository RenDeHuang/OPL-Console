package upstream

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientForwardInjectsBearerAndOperationHeaders(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotAuth string
	var gotIdempotency string
	var gotCorrelation string
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotIdempotency = r.Header.Get("Idempotency-Key")
		gotCorrelation = r.Header.Get("X-Correlation-Id")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := New(ClientConfig{BaseURL: server.URL, BearerToken: "service-token"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/compute-allocations", strings.NewReader(`{"name":"alpha"}`))
	request.Header.Set("content-type", "application/json")

	client.Forward(recorder, request, http.MethodPost, "/api/fabric/compute-allocations")

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if gotMethod != http.MethodPost || gotPath != "/api/fabric/compute-allocations" {
		t.Fatalf("upstream target = %s %s", gotMethod, gotPath)
	}
	if gotAuth != "Bearer service-token" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if gotIdempotency == "" {
		t.Fatalf("Idempotency-Key missing")
	}
	if gotCorrelation == "" {
		t.Fatalf("X-Correlation-Id missing")
	}
	if gotBody != `{"name":"alpha"}` {
		t.Fatalf("body = %q", gotBody)
	}
}

func TestClientForwardReturnsGatewayErrorWhenUpstreamUnavailable(t *testing.T) {
	client := New(ClientConfig{BaseURL: "http://127.0.0.1:1", BearerToken: "service-token"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/compute-allocations", strings.NewReader(`{}`))

	client.Forward(recorder, request, http.MethodPost, "/api/fabric/compute-allocations")

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "upstream_unavailable") {
		t.Fatalf("body = %q", recorder.Body.String())
	}
}
