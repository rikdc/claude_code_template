package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the database connection with additional functionality
type DB struct {
	conn *sql.DB
	path string
}

// Config holds database configuration
type Config struct {
	DatabasePath    string
	MigrationsDir   string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	BusyTimeout     time.Duration
	WALMode         bool
	Synchronous     string
	CacheSize       int
}

// DefaultConfig returns default database configuration optimized for SQLite
func DefaultConfig() *Config {
	return &Config{
		DatabasePath:    "data/prompt_manager.db",
		MigrationsDir:   "database/migrations",
		MaxOpenConns:    1,                    // SQLite works best with single writer
		MaxIdleConns:    1,                    // Keep connection alive
		ConnMaxLifetime: 0,                    // No connection lifetime limit
		ConnMaxIdleTime: 30 * time.Minute,    // Close idle connections after 30 minutes
		BusyTimeout:     30 * time.Second,     // Wait up to 30 seconds for lock
		WALMode:         true,                 // Use WAL mode for better concurrency
		Synchronous:     "NORMAL",             // Balance between safety and performance
		CacheSize:       10000,                // 10MB cache (10000 pages * 1KB)
	}
}

// ProductionConfig returns production-optimized database configuration
func ProductionConfig(dbPath string) *Config {
	return &Config{
		DatabasePath:    dbPath,
		MigrationsDir:   "database/migrations",
		MaxOpenConns:    1,                    // SQLite single writer
		MaxIdleConns:    1,                    // Keep connection alive
		ConnMaxLifetime: 0,                    // No connection lifetime limit
		ConnMaxIdleTime: 10 * time.Minute,    // Shorter idle time in production
		BusyTimeout:     60 * time.Second,     // Longer timeout for production
		WALMode:         true,                 // WAL mode for better performance
		Synchronous:     "NORMAL",             // Good balance for production
		CacheSize:       20000,                // 20MB cache for production
	}
}

// New creates a new database connection with optimized SQLite settings
func New(config *Config) (*DB, error) {
	// Ensure database directory exists
	dir := filepath.Dir(config.DatabasePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Build connection string with SQLite pragmas
	connStr := buildConnectionString(config)

	// Open database connection
	conn, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(config.MaxOpenConns)
	conn.SetMaxIdleConns(config.MaxIdleConns)
	conn.SetConnMaxLifetime(config.ConnMaxLifetime)
	conn.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Apply additional SQLite optimizations
	if err := applySQLiteOptimizations(conn, config); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to apply SQLite optimizations: %w", err)
	}

	db := &DB{
		conn: conn,
		path: config.DatabasePath,
	}

	return db, nil
}

// buildConnectionString constructs SQLite connection string with pragmas
func buildConnectionString(config *Config) string {
	connStr := config.DatabasePath + "?"
	
	// Enable foreign keys
	connStr += "foreign_keys=1"
	
	// Set busy timeout
	if config.BusyTimeout > 0 {
		connStr += fmt.Sprintf("&_busy_timeout=%d", int(config.BusyTimeout/time.Millisecond))
	}
	
	// Enable WAL mode if configured
	if config.WALMode {
		connStr += "&_journal_mode=WAL"
	}
	
	// Set synchronous mode
	if config.Synchronous != "" {
		connStr += fmt.Sprintf("&_sync=%s", config.Synchronous)
	}
	
	return connStr
}

// applySQLiteOptimizations applies runtime SQLite optimizations
func applySQLiteOptimizations(conn *sql.DB, config *Config) error {
	optimizations := []struct {
		pragma string
		value  interface{}
		desc   string
	}{
		{"cache_size", -config.CacheSize, "Set cache size"},
		{"temp_store", "MEMORY", "Store temporary tables in memory"},
		{"mmap_size", 268435456, "Enable memory-mapped I/O (256MB)"},
		{"optimize", nil, "Optimize database"},
	}
	
	for _, opt := range optimizations {
		var query string
		if opt.value != nil {
			query = fmt.Sprintf("PRAGMA %s = %v", opt.pragma, opt.value)
		} else {
			query = fmt.Sprintf("PRAGMA %s", opt.pragma)
		}
		
		if _, err := conn.Exec(query); err != nil {
			return fmt.Errorf("failed to apply %s: %w", opt.desc, err)
		}
	}
	
	return nil
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
		content, err := os.ReadFile(file)
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

// Stats returns database statistics including SQLite-specific metrics
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

	// Connection pool stats
	dbStats := db.conn.Stats()
	stats["connection_pool"] = map[string]interface{}{
		"max_open_connections":     dbStats.MaxOpenConnections,
		"open_connections":         dbStats.OpenConnections,
		"in_use":                  dbStats.InUse,
		"idle":                    dbStats.Idle,
		"wait_count":              dbStats.WaitCount,
		"wait_duration_ms":        dbStats.WaitDuration.Milliseconds(),
		"max_idle_closed":         dbStats.MaxIdleClosed,
		"max_idle_time_closed":    dbStats.MaxIdleTimeClosed,
		"max_lifetime_closed":     dbStats.MaxLifetimeClosed,
	}

	// SQLite-specific stats
	sqliteStats, err := db.getSQLiteStats()
	if err == nil {
		stats["sqlite"] = sqliteStats
	}

	return stats, nil
}

// getSQLiteStats retrieves SQLite-specific statistics and settings
func (db *DB) getSQLiteStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// SQLite pragma values to check
	pragmas := []string{
		"journal_mode",
		"synchronous", 
		"cache_size",
		"temp_store",
		"mmap_size",
		"page_count",
		"page_size",
		"freelist_count",
		"foreign_keys",
	}
	
	for _, pragma := range pragmas {
		var value string
		query := fmt.Sprintf("PRAGMA %s", pragma)
		err := db.conn.QueryRow(query).Scan(&value)
		if err != nil {
			// Some pragmas might not be available, continue
			continue
		}
		stats[pragma] = value
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