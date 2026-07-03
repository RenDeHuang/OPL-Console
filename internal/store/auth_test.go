package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/RenDeHuang/opl-console/internal/auth"
)

func TestAuthStoreFindUserByEmail(t *testing.T) {
	ctx := context.Background()
	pool := authTestPool(ctx, t)
	store := NewAuthStore(pool)
	userID := authTestID(t, "usr")
	passwordHash, err := auth.HashPassword("secret")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	insertAuthUser(ctx, t, pool, userID, "owner@opl.local", passwordHash, "owner", "active")

	user, err := store.FindUserByEmail(ctx, "owner@opl.local")
	if err != nil {
		t.Fatalf("FindUserByEmail: %v", err)
	}

	if user.ID != userID || user.PasswordHash != passwordHash || user.Role != auth.RoleOwner || user.Status != auth.StatusActive {
		t.Fatalf("user = %#v", user)
	}
}

func TestAuthStoreSessionLifecycle(t *testing.T) {
	ctx := context.Background()
	pool := authTestPool(ctx, t)
	store := NewAuthStore(pool)
	userID := authTestID(t, "usr")
	passwordHash, err := auth.HashPassword("secret")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	insertAuthUser(ctx, t, pool, userID, "admin@opl.local", passwordHash, "admin", "active")

	tokenHash := auth.HashToken("plain-token")
	err = store.CreateSession(ctx, auth.SessionRecord{
		ID:            "ignored-session-id",
		UserID:        userID,
		TokenHash:     tokenHash,
		CSRFTokenHash: auth.HashToken("csrf-token"),
		ExpiresAt:     time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	session, err := store.FindSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		t.Fatalf("FindSessionByTokenHash: %v", err)
	}
	if session.User.ID != userID || session.User.Role != auth.RoleAdmin {
		t.Fatalf("session = %#v", session)
	}

	if err := store.DeleteSessionByTokenHash(ctx, tokenHash); err != nil {
		t.Fatalf("DeleteSessionByTokenHash: %v", err)
	}
	_, err = store.FindSessionByTokenHash(ctx, tokenHash)
	if !errors.Is(err, auth.ErrSessionNotFound) {
		t.Fatalf("FindSessionByTokenHash after delete err = %v, want ErrSessionNotFound", err)
	}
}

func TestAuthStoreSeedBootstrapUsersUpsertsUsers(t *testing.T) {
	ctx := context.Background()
	pool := authTestPool(ctx, t)
	store := NewAuthStore(pool)
	userID := authTestID(t, "usr-seed")
	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM sessions WHERE user_id = $1`, userID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM users WHERE id = $1`, userID)
	})

	err := store.SeedBootstrapUsers(ctx, []auth.BootstrapUser{{
		ID:       userID,
		Email:    "seed-owner@opl.local",
		Password: "seed-password",
		Role:     auth.RoleOwner,
		Name:     "Seed Owner",
	}})
	if err != nil {
		t.Fatalf("SeedBootstrapUsers: %v", err)
	}

	user, err := store.FindUserByEmail(ctx, "seed-owner@opl.local")
	if err != nil {
		t.Fatalf("FindUserByEmail: %v", err)
	}
	if !auth.VerifyPassword(user.PasswordHash, "seed-password") {
		t.Fatalf("seeded password hash did not verify")
	}

	err = store.SeedBootstrapUsers(ctx, []auth.BootstrapUser{{
		ID:       userID,
		Email:    "seed-owner@opl.local",
		Password: "seed-password-updated",
		Role:     auth.RoleAdmin,
		Name:     "Seed Admin",
	}})
	if err != nil {
		t.Fatalf("SeedBootstrapUsers update: %v", err)
	}
	user, err = store.FindUserByEmail(ctx, "seed-owner@opl.local")
	if err != nil {
		t.Fatalf("FindUserByEmail updated: %v", err)
	}
	if user.Role != auth.RoleAdmin || !auth.VerifyPassword(user.PasswordHash, "seed-password-updated") {
		t.Fatalf("updated user = %#v", user)
	}
}

func authTestPool(ctx context.Context, t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func insertAuthUser(ctx context.Context, t *testing.T, pool *pgxpool.Pool, id string, email string, passwordHash string, role string, status string) {
	t.Helper()
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, email, name, password_hash, role, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, email, email, passwordHash, role, status)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM sessions WHERE user_id = $1`, id)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM users WHERE id = $1`, id)
	})
}

func authTestID(t *testing.T, prefix string) string {
	t.Helper()
	name := strings.NewReplacer("/", "-", " ", "-", "_", "-").Replace(strings.ToLower(t.Name()))
	return fmt.Sprintf("%s-%s-%d", prefix, name, time.Now().UnixNano())
}
