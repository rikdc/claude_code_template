package handlers

import (
	"path/filepath"
	"testing"

	"github.com/claude-code-template/prompt-manager/internal/database"
)

// setupTestDB creates a temporary test database with migrations applied
func setupTestDB(t *testing.T) *database.DB {
	// Create temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	config := &database.Config{
		DatabasePath:  dbPath,
		MigrationsDir: "../../../database/migrations",
	}
	
	db, err := database.New(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	
	// Run migrations
	if err := db.RunMigrations(config.MigrationsDir); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
	
	return db
}