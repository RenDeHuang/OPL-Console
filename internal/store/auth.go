package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/RenDeHuang/opl-console/internal/auth"
)

type AuthStore struct {
	pool *pgxpool.Pool
}

var (
	_ auth.UserRepository    = (*AuthStore)(nil)
	_ auth.SessionRepository = (*AuthStore)(nil)
)

func NewAuthStore(pool *pgxpool.Pool) *AuthStore {
	return &AuthStore{pool: pool}
}

func (s *AuthStore) FindUserByEmail(ctx context.Context, email string) (auth.UserWithPassword, error) {
	var user auth.UserWithPassword
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, role, status, password_hash
		FROM users
		WHERE lower(email) = lower($1)
	`, email).Scan(&user.ID, &user.Email, &user.Role, &user.Status, &user.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.UserWithPassword{}, auth.ErrInvalidCredentials
	}
	return user, err
}

func (s *AuthStore) CreateSession(ctx context.Context, record auth.SessionRecord) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sessions (id, user_id, csrf_token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
	`, record.TokenHash, record.UserID, record.CSRFTokenHash, record.ExpiresAt)
	return err
}

func (s *AuthStore) FindSessionByTokenHash(ctx context.Context, tokenHash string) (auth.Session, error) {
	var session auth.Session
	err := s.pool.QueryRow(ctx, `
		SELECT s.id, s.csrf_token_hash, s.expires_at, u.id, u.email, u.role, u.status
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.id = $1 AND s.expires_at > now()
	`, tokenHash).Scan(
		&session.Token,
		&session.CSRFToken,
		&session.ExpiresAt,
		&session.User.ID,
		&session.User.Email,
		&session.User.Role,
		&session.User.Status,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.Session{}, auth.ErrSessionNotFound
	}
	return session, err
}

func (s *AuthStore) DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, tokenHash)
	return err
}

func (s *AuthStore) SeedBootstrapUsers(ctx context.Context, users []auth.BootstrapUser) error {
	for _, user := range users {
		passwordHash, err := auth.HashPassword(user.Password)
		if err != nil {
			return err
		}
		if _, err := s.pool.Exec(ctx, `
			INSERT INTO users (id, email, name, password_hash, role, status)
			VALUES ($1, $2, $3, $4, $5, 'active')
			ON CONFLICT (email) DO UPDATE
			SET name = EXCLUDED.name,
			    password_hash = EXCLUDED.password_hash,
			    role = EXCLUDED.role,
			    status = 'active',
			    updated_at = now()
		`, user.ID, user.Email, user.Name, passwordHash, user.Role); err != nil {
			return err
		}
	}
	return nil
}
