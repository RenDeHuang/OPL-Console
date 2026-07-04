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
	request      workspace.CreateWorkspaceRequest
	action       workspace.ActionRequest
	token        workspace.TokenRequest
	result       workspace.CreateWorkspaceResult
	actionResult workspace.ActionResult
	handoff      workspace.HandoffResult
	err          error
}

func (f *fakeWorkspaceService) CreateWorkspace(ctx context.Context, request workspace.CreateWorkspaceRequest) (workspace.CreateWorkspaceResult, error) {
	f.request = request
	if f.err != nil {
		return workspace.CreateWorkspaceResult{}, f.err
	}
	return f.result, nil
}

func (f *fakeWorkspaceService) ConfigureWorkspace(ctx context.Context, request workspace.ActionRequest) (workspace.ActionResult, error) {
	f.action = request
	return f.actionResult, f.err
}

func (f *fakeWorkspaceService) SuspendWorkspace(ctx context.Context, request workspace.ActionRequest) (workspace.ActionResult, error) {
	f.action = request
	return f.actionResult, f.err
}

func (f *fakeWorkspaceService) DeleteWorkspace(ctx context.Context, request workspace.ActionRequest) (workspace.ActionResult, error) {
	f.action = request
	return f.actionResult, f.err
}

func (f *fakeWorkspaceService) ResetWorkspaceToken(ctx context.Context, request workspace.TokenRequest) (workspace.ActionResult, error) {
	f.token = request
	return f.actionResult, f.err
}

func (f *fakeWorkspaceService) DeleteWorkspaceToken(ctx context.Context, request workspace.ActionRequest) (workspace.ActionResult, error) {
	f.action = request
	return f.actionResult, f.err
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
	request.Header.Set("x-opl-csrf-token", "csrf-token")
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

func TestResetWorkspaceTokenRouteReadsToken(t *testing.T) {
	workspaceService := &fakeWorkspaceService{actionResult: workspace.ActionResult{WorkspaceID: "ws-alpha", State: "ready", URL: "https://workspace.example/ws-alpha"}}
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
	request := httptest.NewRequest(http.MethodPost, "/api/workspaces/reset-token", strings.NewReader(`{"workspaceId":"ws-alpha","token":"new-token"}`))
	request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
	request.Header.Set("x-opl-csrf-token", "csrf-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if workspaceService.token.WorkspaceID != "ws-alpha" || workspaceService.token.Token != "new-token" || workspaceService.token.ActorUserID != "usr-owner" {
		t.Fatalf("token request = %#v", workspaceService.token)
	}
}

func TestDeleteWorkspaceTokenRouteReadsWorkspaceIDFromBody(t *testing.T) {
	workspaceService := &fakeWorkspaceService{actionResult: workspace.ActionResult{WorkspaceID: "ws-alpha", State: "token_deleted"}}
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
	request := httptest.NewRequest(http.MethodPost, "/api/workspaces/delete-token", strings.NewReader(`{"workspaceId":"ws-alpha"}`))
	request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
	request.Header.Set("x-opl-csrf-token", "csrf-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if workspaceService.action.WorkspaceID != "ws-alpha" || workspaceService.action.ActorUserID != "usr-owner" {
		t.Fatalf("action = %#v", workspaceService.action)
	}
}

func TestOldWorkspaceCompatibilityRoutesAreNotMounted(t *testing.T) {
	handler := NewRouter(Dependencies{})
	for _, path := range []string{
		"/api/workspaces/ws-alpha/configure",
		"/api/workspaces/ws-alpha/suspend",
		"/api/workspaces/ws-alpha/delete",
		"/api/workspaces/ws-alpha/tokens/reset",
		"/api/workspaces/ws-alpha/tokens/delete",
	} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusNotFound {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}
}
