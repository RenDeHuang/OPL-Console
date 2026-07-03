package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/RenDeHuang/opl-console/internal/auth"
)

const defaultSessionCookieName = "opl_console_session"

type AuthService interface {
	Login(ctx context.Context, email string, password string) (auth.Session, error)
	Session(ctx context.Context, token string) (auth.Session, error)
	Logout(ctx context.Context, token string) error
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	User      auth.User `json:"user"`
	CSRFToken string    `json:"csrfToken"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func mountAuthRoutes(router Router, deps Dependencies) {
	cookieName := deps.SessionCookieName
	if cookieName == "" {
		cookieName = defaultSessionCookieName
	}

	router.Post("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if deps.Auth == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "auth_not_configured"})
			return
		}
		var payload authRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}
		session, err := deps.Auth.Login(r.Context(), payload.Email, payload.Password)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) || errors.Is(err, auth.ErrUserDisabled) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_credentials"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "auth_failed"})
			return
		}
		setSessionCookie(w, cookieName, session.Token, session.ExpiresAt)
		writeAuthResponse(w, session)
	})

	router.Get("/api/auth/session", func(w http.ResponseWriter, r *http.Request) {
		if deps.Auth == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not_authenticated"})
			return
		}
		cookie, err := r.Cookie(cookieName)
		if err != nil || cookie.Value == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not_authenticated"})
			return
		}
		session, err := deps.Auth.Session(r.Context(), cookie.Value)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not_authenticated"})
			return
		}
		writeAuthResponse(w, session)
	})

	router.Post("/api/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		if deps.Auth != nil {
			if cookie, err := r.Cookie(cookieName); err == nil && cookie.Value != "" {
				_ = deps.Auth.Logout(r.Context(), cookie.Value)
			}
		}
		clearSessionCookie(w, cookieName)
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
}

func writeAuthResponse(w http.ResponseWriter, session auth.Session) {
	writeJSON(w, http.StatusOK, authResponse{
		User:      session.User,
		CSRFToken: session.CSRFToken,
		ExpiresAt: session.ExpiresAt,
	})
}

func setSessionCookie(w http.ResponseWriter, name string, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
