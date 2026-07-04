package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeUserRepository struct {
	user UserWithPassword
	err  error
}

func (f fakeUserRepository) FindUserByEmail(ctx context.Context, email string) (UserWithPassword, error) {
	if f.err != nil {
		return UserWithPassword{}, f.err
	}
	if f.user.Email != email {
		return UserWithPassword{}, ErrInvalidCredentials
	}
	return f.user, nil
}

func (f fakeUserRepository) FindFirstAdmin(ctx context.Context) (UserWithPassword, error) {
	if f.err != nil {
		return UserWithPassword{}, f.err
	}
	if f.user.Role != RoleAdmin || f.user.Status != StatusActive {
		return UserWithPassword{}, ErrInvalidCredentials
	}
	return f.user, nil
}

type fakeSessionRepository struct {
	createdSession SessionRecord
	lookupSession  Session
	lookupErr      error
	deletedHash    string
}

func (f *fakeSessionRepository) CreateSession(ctx context.Context, record SessionRecord) error {
	f.createdSession = record
	return nil
}

func (f *fakeSessionRepository) FindSessionByTokenHash(ctx context.Context, tokenHash string) (Session, error) {
	if f.lookupErr != nil {
		return Session{}, f.lookupErr
	}
	return f.lookupSession, nil
}

func (f *fakeSessionRepository) DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error {
	f.deletedHash = tokenHash
	return nil
}

func TestServiceLoginCreatesSessionForActiveUser(t *testing.T) {
	passwordHash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	sessionRepo := &fakeSessionRepository{}
	service := NewService(ServiceConfig{
		Users: fakeUserRepository{user: UserWithPassword{
			User: User{
				ID:     "usr-owner",
				Email:  "owner@opl.local",
				Role:   RoleOwner,
				Status: StatusActive,
			},
			PasswordHash: passwordHash,
		}},
		Sessions: sessionRepo,
		Now:      func() time.Time { return time.Unix(1000, 0).UTC() },
	})

	session, err := service.Login(context.Background(), "owner@opl.local", "correct-password")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if session.Token == "" || session.CSRFToken == "" {
		t.Fatalf("session tokens must be generated: %#v", session)
	}
	if sessionRepo.createdSession.TokenHash == "" || sessionRepo.createdSession.CSRFTokenHash == "" {
		t.Fatalf("hashed tokens must be persisted: %#v", sessionRepo.createdSession)
	}
	if sessionRepo.createdSession.TokenHash == session.Token {
		t.Fatalf("session token must be hashed before persistence")
	}
	if session.ExpiresAt.Sub(time.Unix(1000, 0).UTC()) != 24*time.Hour {
		t.Fatalf("expiresAt = %s", session.ExpiresAt)
	}
}

func TestServiceLoginRejectsBadPasswordAndDisabledUser(t *testing.T) {
	passwordHash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	tests := []struct {
		name       string
		status     Status
		password   string
		wantErr    error
		createSess bool
	}{
		{name: "bad password", status: StatusActive, password: "bad-password", wantErr: ErrInvalidCredentials},
		{name: "disabled", status: StatusDisabled, password: "correct-password", wantErr: ErrUserDisabled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := &fakeSessionRepository{}
			service := NewService(ServiceConfig{
				Users: fakeUserRepository{user: UserWithPassword{
					User: User{
						ID:     "usr-owner",
						Email:  "owner@opl.local",
						Role:   RoleOwner,
						Status: tt.status,
					},
					PasswordHash: passwordHash,
				}},
				Sessions: sessionRepo,
			})

			_, err := service.Login(context.Background(), "owner@opl.local", tt.password)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if sessionRepo.createdSession.TokenHash != "" {
				t.Fatalf("session should not be created: %#v", sessionRepo.createdSession)
			}
		})
	}
}

func TestServiceSessionAndLogoutHashTokenBeforeRepositoryUse(t *testing.T) {
	sessionRepo := &fakeSessionRepository{lookupSession: Session{
		Token:     "stored-hash",
		CSRFToken: "stored-csrf",
		User:      User{ID: "usr-admin", Email: "admin@opl.local", Role: RoleAdmin, Status: StatusActive},
		ExpiresAt: time.Now().Add(time.Hour),
	}}
	service := NewService(ServiceConfig{
		Users:    fakeUserRepository{},
		Sessions: sessionRepo,
	})

	session, err := service.Session(context.Background(), "plain-token")
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	if session.User.ID != "usr-admin" {
		t.Fatalf("session user = %#v", session.User)
	}
	if err := service.Logout(context.Background(), "plain-token"); err != nil {
		t.Fatalf("logout: %v", err)
	}
	expectedHash := HashToken("plain-token")
	if sessionRepo.deletedHash != expectedHash {
		t.Fatalf("deletedHash = %q, want %q", sessionRepo.deletedHash, expectedHash)
	}
}
