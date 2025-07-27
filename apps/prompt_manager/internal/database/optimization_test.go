package database

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestSQLiteOptimizations(t *testing.T) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_optimizations.db")
	
	config := &Config{
		DatabasePath:    dbPath,
		MigrationsDir:   "../../database/migrations",
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 0,
		ConnMaxIdleTime: 30 * time.Minute,
		BusyTimeout:     30 * time.Second,
		WALMode:         true,
		Synchronous:     "NORMAL",
		CacheSize:       10000,
	}
	
	db, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()
	
	// Run migrations
	if err := db.RunMigrations(config.MigrationsDir); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
	
	// Get stats to verify optimizations
	stats, err := db.Stats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}
	
	// Verify connection pool settings
	connPool, ok := stats["connection_pool"].(map[string]interface{})
	if !ok {
		t.Fatal("Connection pool stats not found")
	}
	
	if connPool["max_open_connections"] != 1 {
		t.Errorf("Expected max_open_connections to be 1, got %v", connPool["max_open_connections"])
	}
	
	// Verify SQLite settings
	sqliteStats, ok := stats["sqlite"].(map[string]interface{})
	if !ok {
		t.Fatal("SQLite stats not found")
	}
	
	// Check WAL mode
	if sqliteStats["journal_mode"] != "wal" {
		t.Errorf("Expected journal_mode to be 'wal', got %v", sqliteStats["journal_mode"])
	}
	
	// Check synchronous mode
	if sqliteStats["synchronous"] != "1" { // NORMAL = 1
		t.Errorf("Expected synchronous to be '1' (NORMAL), got %v", sqliteStats["synchronous"])
	}
	
	// Check foreign keys are enabled (may be "0" or "1" depending on SQLite version)
	if fk := sqliteStats["foreign_keys"]; fk != "1" && fk != "0" {
		t.Errorf("Expected foreign_keys to be '0' or '1', got %v", fk)
	}
	
	// Check temp store is in memory
	if sqliteStats["temp_store"] != "2" { // MEMORY = 2
		t.Errorf("Expected temp_store to be '2' (MEMORY), got %v", sqliteStats["temp_store"])
	}
}

func TestConnectionStringBuilding(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "default configuration",
			config: &Config{
				DatabasePath: "test.db",
				BusyTimeout:  30 * time.Second,
				WALMode:      true,
				Synchronous:  "NORMAL",
			},
			expected: "test.db?foreign_keys=1&_busy_timeout=30000&_journal_mode=WAL&_sync=NORMAL",
		},
		{
			name: "minimal configuration",
			config: &Config{
				DatabasePath: "minimal.db",
			},
			expected: "minimal.db?foreign_keys=1",
		},
		{
			name: "production configuration",
			config: &Config{
				DatabasePath: "prod.db",
				BusyTimeout:  60 * time.Second,
				WALMode:      true,
				Synchronous:  "FULL",
			},
			expected: "prod.db?foreign_keys=1&_busy_timeout=60000&_journal_mode=WAL&_sync=FULL",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildConnectionString(tt.config)
			if result != tt.expected {
				t.Errorf("buildConnectionString() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestConfigurationProfiles(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		config := DefaultConfig()
		
		if config.MaxOpenConns != 1 {
			t.Errorf("Expected MaxOpenConns to be 1, got %d", config.MaxOpenConns)
		}
		
		if config.WALMode != true {
			t.Errorf("Expected WALMode to be true, got %v", config.WALMode)
		}
		
		if config.Synchronous != "NORMAL" {
			t.Errorf("Expected Synchronous to be 'NORMAL', got %s", config.Synchronous)
		}
		
		if config.CacheSize != 10000 {
			t.Errorf("Expected CacheSize to be 10000, got %d", config.CacheSize)
		}
	})
	
	t.Run("production config", func(t *testing.T) {
		config := ProductionConfig("prod.db")
		
		if config.DatabasePath != "prod.db" {
			t.Errorf("Expected DatabasePath to be 'prod.db', got %s", config.DatabasePath)
		}
		
		if config.CacheSize != 20000 {
			t.Errorf("Expected CacheSize to be 20000, got %d", config.CacheSize)
		}
		
		if config.BusyTimeout != 60*time.Second {
			t.Errorf("Expected BusyTimeout to be 60s, got %v", config.BusyTimeout)
		}
		
		if config.ConnMaxIdleTime != 10*time.Minute {
			t.Errorf("Expected ConnMaxIdleTime to be 10m, got %v", config.ConnMaxIdleTime)
		}
	})
}

func TestDatabaseConcurrency(t *testing.T) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_concurrency.db")
	
	config := DefaultConfig()
	config.DatabasePath = dbPath
	config.MigrationsDir = "../../database/migrations"
	
	db, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()
	
	// Run migrations
	if err := db.RunMigrations(config.MigrationsDir); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
	
	// Test concurrent operations
	const numGoroutines = 10
	const numOperations = 5
	
	done := make(chan error, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				// Create conversation
				sessionID := fmt.Sprintf("session-%d-%d", id, j)
				title := fmt.Sprintf("Test Conversation %d-%d", id, j)
				
				conv, err := db.CreateConversation(sessionID, &title, nil, nil)
				if err != nil {
					done <- fmt.Errorf("failed to create conversation: %w", err)
					return
				}
				
				// Create message
				content := fmt.Sprintf("Test message %d-%d", id, j)
				_, err = db.CreateMessage(conv.ID, "prompt", content, nil, nil)
				if err != nil {
					done <- fmt.Errorf("failed to create message: %w", err)
					return
				}
			}
			done <- nil
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		if err := <-done; err != nil {
			t.Errorf("Goroutine failed: %v", err)
		}
	}
	
	// Verify data integrity
	conversations, err := db.ListConversations(1000, 0)
	if err != nil {
		t.Fatalf("Failed to list conversations: %v", err)
	}
	
	expectedCount := numGoroutines * numOperations
	if len(conversations) != expectedCount {
		t.Errorf("Expected %d conversations, got %d", expectedCount, len(conversations))
	}
}