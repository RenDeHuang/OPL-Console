package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RenDeHuang/opl-console/internal/auth"
	"github.com/RenDeHuang/opl-console/internal/console"
)

type fakeGovernanceService struct {
	me         console.Me
	packages   []console.Package
	workspaces []console.ManagedWorkspace
	adminUsers []console.UserView
}

func (f fakeGovernanceService) Me(ctx context.Context, user auth.User) (console.Me, error) {
	return f.me, nil
}

func (f fakeGovernanceService) Packages(ctx context.Context) ([]console.Package, error) {
	return f.packages, nil
}

func (f fakeGovernanceService) Workspaces(ctx context.Context, user auth.User) ([]console.ManagedWorkspace, error) {
	return f.workspaces, nil
}

func (f fakeGovernanceService) AdminUsers(ctx context.Context) ([]console.UserView, error) {
	return f.adminUsers, nil
}

func TestOwnerGovernanceRoutesRequireActiveOwnerSession(t *testing.T) {
	authService := &fakeAuthService{session: auth.Session{
		Token:     "session-token",
		CSRFToken: "csrf-token",
		ExpiresAt: time.Now().Add(time.Hour),
		User: auth.User{
			ID:     "usr-owner",
			Email:  "owner@opl.local",
			Role:   auth.RoleOwner,
			Status: auth.StatusActive,
		},
	}}
	handler := NewRouter(Dependencies{
		Auth:              authService,
		SessionCookieName: "opl_session",
		Governance: fakeGovernanceService{
			me: console.Me{
				User:         console.UserView{ID: "usr-owner", Email: "owner@opl.local", Role: "owner", Status: "active"},
				Organization: console.OrganizationView{ID: "org-alpha", Name: "Alpha Lab", Status: "active"},
			},
			packages: []console.Package{{ID: "basic", Name: "Basic Workspace", CPU: 2, MemoryGB: 4, StorageGB: 10}},
			workspaces: []console.ManagedWorkspace{{
				ID:     "ws-alpha",
				Name:   "Alpha Workspace",
				State:  "running",
				Policy: "managed",
			}},
		},
	})

	for _, path := range []string{"/api/me", "/api/packages", "/api/workspaces"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestOwnerGovernanceRoutesRejectMissingSession(t *testing.T) {
	handler := NewRouter(Dependencies{
		Auth:              &fakeAuthService{},
		SessionCookieName: "opl_session",
		Governance:        fakeGovernanceService{},
	})
	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestAdminGovernanceRoutesRequireAdmin(t *testing.T) {
	tests := []struct {
		name string
		user auth.User
		want int
	}{
		{
			name: "admin",
			user: auth.User{ID: "usr-admin", Email: "admin@opl.local", Role: auth.RoleAdmin, Status: auth.StatusActive},
			want: http.StatusOK,
		},
		{
			name: "owner",
			user: auth.User{ID: "usr-owner", Email: "owner@opl.local", Role: auth.RoleOwner, Status: auth.StatusActive},
			want: http.StatusForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewRouter(Dependencies{
				Auth: &fakeAuthService{session: auth.Session{
					Token:     "session-token",
					CSRFToken: "csrf-token",
					ExpiresAt: time.Now().Add(time.Hour),
					User:      tt.user,
				}},
				SessionCookieName: "opl_session",
				Governance: fakeGovernanceService{
					adminUsers: []console.UserView{{ID: "usr-admin", Email: "admin@opl.local", Role: "admin", Status: "active"}},
				},
			})
			request := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
			request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d, body = %s", response.Code, tt.want, response.Body.String())
			}
			if response.Code == http.StatusOK {
				var users []console.UserView
				if err := json.NewDecoder(response.Body).Decode(&users); err != nil {
					t.Fatalf("decode users: %v", err)
				}
				if len(users) != 1 {
					t.Fatalf("users = %#v", users)
				}
			}
		})
	}
}
