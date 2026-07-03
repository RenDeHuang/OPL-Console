package store

import (
	"context"
	"os"
	"testing"
)

func TestOpenRequiresReachableDatabase(t *testing.T) {
	t.Setenv("PGCONNECT_TIMEOUT", "1")
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	pool, err := Open(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer pool.Close()
}
