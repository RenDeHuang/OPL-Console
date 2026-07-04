package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RenDeHuang/opl-console/internal/auth"
)

type fakeAuthService struct {
	loginSession    auth.Session
	operatorSession auth.Session
	session         auth.Session
	loginErr        error
	operatorErr     error
	sessionErr      error
	logoutToken     string
}

func (f *fakeAuthService) Login(ctx context.Context, email string, password string) (auth.Session, error) {
	if f.loginErr != nil {
		return auth.Session{}, f.loginErr
	}
	return f.loginSession, nil
}

func (f *fakeAuthService) OperatorLogin(ctx context.Context) (auth.Session, error) {
	if f.operatorErr != nil {
		return auth.Session{}, f.operatorErr
	}
	return f.operatorSession, nil
}

func (f *fakeAuthService) Session(ctx context.Context, token string) (auth.Session, error) {
	if f.sessionErr != nil {
		return auth.Session{}, f.sessionErr
	}
	return f.session, nil
}

func (f *fakeAuthService) Logout(ctx context.Context, token string) error {
	f.logoutToken = token
	return nil
}

func TestAuthLoginSetsSessionCookieWithoutLeakingPassword(t *testing.T) {
	service := &fakeAuthService{loginSession: auth.Session{
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
	handler := NewRouter(Dependencies{Auth: service, SessionCookieName: "opl_session"})
	body := bytes.NewBufferString(`{"email":"owner@opl.local","password":"secret"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	cookie := response.Result().Cookies()[0]
	if cookie.Name != "opl_session" || cookie.Value != "session-token" {
		t.Fatalf("cookie = %#v", cookie)
	}
	if !cookie.HttpOnly {
		t.Fatalf("session cookie must be HttpOnly")
	}
	if bytes.Contains(response.Body.Bytes(), []byte("secret")) {
		t.Fatalf("login response leaked password: %s", response.Body.String())
	}

	var payload authResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.CSRFToken != "csrf-token" {
		t.Fatalf("csrfToken = %q", payload.CSRFToken)
	}
	if payload.User.Email != "owner@opl.local" {
		t.Fatalf("user.email = %q", payload.User.Email)
	}
}

func TestAuthLoginRejectsInvalidCredentials(t *testing.T) {
	handler := NewRouter(Dependencies{
		Auth:              &fakeAuthService{loginErr: auth.ErrInvalidCredentials},
		SessionCookieName: "opl_session",
	})
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"email":"owner@opl.local","password":"bad"}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if cookies := response.Result().Cookies(); len(cookies) != 0 {
		t.Fatalf("unexpected cookies = %#v", cookies)
	}
}

func TestAuthMeReadsSessionCookie(t *testing.T) {
	service := &fakeAuthService{session: auth.Session{
		Token:     "session-token",
		CSRFToken: "csrf-token",
		ExpiresAt: time.Now().Add(time.Hour),
		User: auth.User{
			ID:     "usr-admin",
			Email:  "admin@opl.local",
			Role:   auth.RoleAdmin,
			Status: auth.StatusActive,
		},
	}}
	handler := NewRouter(Dependencies{Auth: service, SessionCookieName: "opl_session"})
	request := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var payload authResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.User.Role != auth.RoleAdmin {
		t.Fatalf("role = %q", payload.User.Role)
	}
}

func TestAuthMeRejectsMissingOrUnknownCookie(t *testing.T) {
	tests := []struct {
		name    string
		cookie  *http.Cookie
		service *fakeAuthService
	}{
		{name: "missing cookie", service: &fakeAuthService{}},
		{
			name:    "unknown cookie",
			cookie:  &http.Cookie{Name: "opl_session", Value: "bad"},
			service: &fakeAuthService{sessionErr: auth.ErrSessionNotFound},
		},
		{
			name:    "store error",
			cookie:  &http.Cookie{Name: "opl_session", Value: "token"},
			service: &fakeAuthService{sessionErr: errors.New("store failed")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewRouter(Dependencies{Auth: tt.service, SessionCookieName: "opl_session"})
			request := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
			if tt.cookie != nil {
				request.AddCookie(tt.cookie)
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestOperatorLoginRequiresConfiguredTokenAndCreatesAdminSession(t *testing.T) {
	service := &fakeAuthService{operatorSession: auth.Session{
		Token:     "operator-session",
		CSRFToken: "operator-csrf",
		ExpiresAt: time.Now().Add(time.Hour),
		User:      auth.User{ID: "usr-admin", Email: "admin@opl.local", Role: auth.RoleAdmin, Status: auth.StatusActive},
	}}
	handler := NewRouter(Dependencies{
		Auth:                 service,
		SessionCookieName:    "opl_session",
		OperatorSummaryToken: "operator-secret",
	})
	request := httptest.NewRequest(http.MethodPost, "/api/auth/operator-login", bytes.NewBufferString(`{"operatorToken":"operator-secret"}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var payload authResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.User.Role != auth.RoleAdmin || payload.CSRFToken != "operator-csrf" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestAuthLogoutClearsSessionCookie(t *testing.T) {
	service := &fakeAuthService{}
	handler := NewRouter(Dependencies{Auth: service, SessionCookieName: "opl_session"})
	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	request.AddCookie(&http.Cookie{Name: "opl_session", Value: "session-token"})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if service.logoutToken != "session-token" {
		t.Fatalf("logout token = %q", service.logoutToken)
	}
	cookie := response.Result().Cookies()[0]
	if cookie.Name != "opl_session" || cookie.MaxAge >= 0 {
		t.Fatalf("clear cookie = %#v", cookie)
	}
}
