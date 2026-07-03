package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitialMigrationIncludesGovernanceTables(t *testing.T) {
	root := filepath.Join("..", "..")
	matches, err := filepath.Glob(filepath.Join(root, "migrations", "*.sql"))
	if err != nil {
		t.Fatalf("glob migrations: %v", err)
	}
	var builder strings.Builder
	for _, migrationPath := range matches {
		content, err := os.ReadFile(migrationPath)
		if err != nil {
			t.Fatalf("read migration %s: %v", migrationPath, err)
		}
		builder.Write(content)
		builder.WriteByte('\n')
	}
	sql := builder.String()

	for _, table := range []string{
		"teams",
		"roles",
		"policies",
		"approvals",
		"managed_resource_views",
	} {
		needle := "CREATE TABLE " + table + " "
		if !strings.Contains(sql, needle) {
			t.Fatalf("initial migration missing %q", needle)
		}
	}
}
