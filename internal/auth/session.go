package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const defaultSessionTTL = 24 * time.Hour

type UserWithPassword struct {
	User
	PasswordHash string
}

type SessionRecord struct {
	ID            string
	UserID        string
	TokenHash     string
	CSRFTokenHash string
	ExpiresAt     time.Time
}

type UserRepository interface {
	FindUserByEmail(ctx context.Context, email string) (UserWithPassword, error)
}

type SessionRepository interface {
	CreateSession(ctx context.Context, record SessionRecord) error
	FindSessionByTokenHash(ctx context.Context, tokenHash string) (Session, error)
	DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error
}

type ServiceConfig struct {
	Users      UserRepository
	Sessions   SessionRepository
	SessionTTL time.Duration
	Now        func() time.Time
}

type Service struct {
	users      UserRepository
	sessions   SessionRepository
	sessionTTL time.Duration
	now        func() time.Time
}

func NewService(cfg ServiceConfig) *Service {
	ttl := cfg.SessionTTL
	if ttl == 0 {
		ttl = defaultSessionTTL
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	return &Service{users: cfg.Users, sessions: cfg.Sessions, sessionTTL: ttl, now: now}
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(passwordHash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) == nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *Service) Login(ctx context.Context, email string, password string) (Session, error) {
	user, err := s.users.FindUserByEmail(ctx, email)
	if err != nil {
		return Session{}, ErrInvalidCredentials
	}
	if user.Status != StatusActive {
		return Session{}, ErrUserDisabled
	}
	if !VerifyPassword(user.PasswordHash, password) {
		return Session{}, ErrInvalidCredentials
	}

	token, err := randomToken()
	if err != nil {
		return Session{}, err
	}
	csrfToken, err := randomToken()
	if err != nil {
		return Session{}, err
	}
	sessionID, err := randomToken()
	if err != nil {
		return Session{}, err
	}
	expiresAt := s.now().Add(s.sessionTTL)
	if err := s.sessions.CreateSession(ctx, SessionRecord{
		ID:            sessionID,
		UserID:        user.ID,
		TokenHash:     HashToken(token),
		CSRFTokenHash: HashToken(csrfToken),
		ExpiresAt:     expiresAt,
	}); err != nil {
		return Session{}, err
	}
	return Session{
		Token:     token,
		CSRFToken: csrfToken,
		User:      user.User,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Service) Session(ctx context.Context, token string) (Session, error) {
	if token == "" {
		return Session{}, ErrSessionNotFound
	}
	return s.sessions.FindSessionByTokenHash(ctx, HashToken(token))
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.sessions.DeleteSessionByTokenHash(ctx, HashToken(token))
}

func randomToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
