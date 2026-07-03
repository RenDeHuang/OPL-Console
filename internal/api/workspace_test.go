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
	"github.com/RenDeHuang/opl-console/internal/workspace"
)

type fakeWorkspaceService struct {
	request workspace.CreateWorkspaceRequest
	result  workspace.CreateWorkspaceResult
	handoff workspace.HandoffResult
	err     error
}

func (f *fakeWorkspaceService) CreateWorkspace(ctx context.Context, request workspace.CreateWorkspaceRequest) (workspace.CreateWorkspaceResult, error) {
	f.request = request
	if f.err != nil {
		return workspace.CreateWorkspaceResult{}, f.err
	}
	return f.result, nil
}

func (f *fakeWorkspaceService) Handoff(ctx context.Context, request workspace.HandoffRequest) (workspace.HandoffResult, error) {
	return f.handoff, f.err
}

func TestCreateWorkspaceRouteRequiresOwnerAndCallsFacade(t *testing.T) {
	workspaceService := &fakeWorkspaceService{result: workspace.CreateWorkspaceResult{
		WorkspaceID: "ws-alpha",
		URL:         "https://workspace.example.com/w/ws-alpha?token=share",
	}}
	handler := NewRouter(Dependencies{
		Auth: &fakeAuthService{session: auth.Session{
			Token:     "session-token",
			CSRFToken: "csrf-token",
			ExpiresAt: time.Now().Add(time.Hour),
			User:      auth.User{ID: "usr-owner", Email: "owner@opl.local", Role: auth.RoleOwner, Status: auth.StatusActive},
		}},
		SessionCookieName: "opl_session",
		Workspace:         workspaceService,
	})
	request := httptest.NewRequest(http.MethodPost, "/api/workspaces", strings.NewReader(`{
		"workspaceId":"ws-alpha",
		"name":"Alpha Workspace",
		"billingAccountId":"billing-alpha",
		"packageId":"basic",
		"token":"share"
	}`))
	request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if workspaceService.request.WorkspaceID != "ws-alpha" || workspaceService.request.PackageID != "basic" {
		t.Fatalf("request = %#v", workspaceService.request)
	}
	var payload workspace.CreateWorkspaceResult
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.URL == "" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestCreateWorkspaceRouteRejectsMissingSession(t *testing.T) {
	handler := NewRouter(Dependencies{Auth: &fakeAuthService{}, Workspace: &fakeWorkspaceService{}})
	request := httptest.NewRequest(http.MethodPost, "/api/workspaces", strings.NewReader(`{}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestWorkspaceHandoffValidatesTokenWithoutSession(t *testing.T) {
	workspaceService := &fakeWorkspaceService{handoff: workspace.HandoffResult{
		WorkspaceID: "ws-alpha",
		URL:         "https://workspace.example/ws-alpha",
		State:       "running",
	}}
	handler := NewRouter(Dependencies{Workspace: workspaceService})
	request := httptest.NewRequest(http.MethodGet, "/w/ws-alpha?token=share-token", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var payload workspace.HandoffResult
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.URL != "https://workspace.example/ws-alpha" {
		t.Fatalf("payload = %#v", payload)
	}
}
