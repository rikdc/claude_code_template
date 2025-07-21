package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the database connection with additional functionality
type DB struct {
	conn *sql.DB
	path string
}

// Config holds database configuration
type Config struct {
	DatabasePath string
	MigrationsDir string
}

// DefaultConfig returns default database configuration
func DefaultConfig() *Config {
	return &Config{
		DatabasePath:  ".claude/apps/prompt_manager/database/prompts.db",
		MigrationsDir: "database/migrations",
	}
}

// New creates a new database connection
func New(config *Config) (*DB, error) {
	// Ensure database directory exists
	dir := filepath.Dir(config.DatabasePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	conn, err := sql.Open("sqlite3", config.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure SQLite connection
	conn.SetMaxOpenConns(1) // SQLite works best with single connection
	conn.SetMaxIdleConns(1)

	// Test connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	db := &DB{
		conn: conn,
		path: config.DatabasePath,
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Conn returns the underlying sql.DB connection
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// RunMigrations executes database migrations from the migrations directory
func (db *DB) RunMigrations(migrationsDir string) error {
	// Create migrations table if it doesn't exist
	createMigrationsTable := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	
	if _, err := db.conn.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Find migration files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to find migration files: %w", err)
	}

	for _, file := range files {
		version := extractVersionFromFilename(file)
		
		// Check if migration already applied
		var count int
		err := db.conn.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}
		
		if count > 0 {
			continue // Skip already applied migration
		}

		// Read and execute migration
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		tx, err := db.conn.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin migration transaction: %w", err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		// Mark migration as applied
		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", file, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", file, err)
		}

		fmt.Printf("Applied migration: %s\n", version)
	}

	return nil
}

// Health checks database connectivity and returns status
func (db *DB) Health() error {
	if db.conn == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	return db.conn.Ping()
}

// Stats returns database statistics
func (db *DB) Stats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Count conversations
	var conversationCount int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM conversations").Scan(&conversationCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count conversations: %w", err)
	}
	stats["conversations"] = conversationCount

	// Count messages
	var messageCount int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM messages").Scan(&messageCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count messages: %w", err)
	}
	stats["messages"] = messageCount

	// Count ratings
	var ratingCount int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM ratings").Scan(&ratingCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count ratings: %w", err)
	}
	stats["ratings"] = ratingCount

	// Database file size
	if info, err := os.Stat(db.path); err == nil {
		stats["database_size_bytes"] = info.Size()
	}

	return stats, nil
}

// extractVersionFromFilename extracts version number from migration filename
// e.g., "001_initial_schema.up.sql" -> "001"
func extractVersionFromFilename(filename string) string {
	base := filepath.Base(filename)
	if len(base) >= 3 {
		return base[:3]
	}
	return base
}