package main

import (
	"os"
	"testing"

	"github.com/claude-code-template/prompt-manager/internal/api"
	"github.com/claude-code-template/prompt-manager/internal/database"
)

func TestMainIntegration(t *testing.T) {
	// Test that we can initialize all components without errors
	
	// Create temp database file
	tmpfile, err := os.CreateTemp("", "test_main_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	// Test database initialization
	config := &database.Config{
		DatabasePath:  tmpfile.Name(),
		MigrationsDir: "../database/migrations",
	}

	db, err := database.New(config)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Test migrations
	if err := db.RunMigrations(config.MigrationsDir); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Test API server initialization
	server := api.NewServer(db)
	if server == nil {
		t.Fatal("Failed to create API server")
	}

	// Test database health
	if err := db.Health(); err != nil {
		t.Errorf("Database health check failed: %v", err)
	}

	// Test database stats
	stats, err := db.Stats()
	if err != nil {
		t.Errorf("Failed to get database stats: %v", err)
	}

	if stats == nil {
		t.Error("Expected non-nil stats")
	}
}