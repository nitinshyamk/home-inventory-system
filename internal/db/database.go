package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultDBPath = "./inventory.db"
	envDBPath     = "INVENTORY_DB_PATH"
)

// Config holds database configuration
type Config struct {
	Path string
}

// DefaultConfig returns the default database configuration
func DefaultConfig() Config {
	path := os.Getenv(envDBPath)
	if path == "" {
		path = defaultDBPath
	}
	return Config{Path: path}
}

// NewConfigWithPath creates a config with the specified path
func NewConfigWithPath(path string) Config {
	if path == "" {
		return DefaultConfig()
	}
	return Config{Path: path}
}

// Open opens a connection to the SQLite database
func Open(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
