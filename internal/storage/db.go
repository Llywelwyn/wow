package storage

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// schema defines the SQL schema for the snippets table.
const schema = `
CREATE TABLE IF NOT EXISTS snippets (
    key TEXT PRIMARY KEY,
    type TEXT NOT NULL DEFAULT 'text',
    created DATETIME NOT NULL,
    modified DATETIME NOT NULL,
    description TEXT,
    tags TEXT
);
`

// InitMetaDB initializes a SQLite database at the given path.
// It ensures the directory exists, opens the database, applies migrations,
// and returns the database handle.
func InitMetaDB(path string) (*sql.DB, error) {
	if err := ensureDir(path); err != nil {
		return nil, err
	}

	dsn := buildDSN(path)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db %q: %w", path, err)
	}

	db.SetMaxOpenConns(1)
	db.SetConnMaxIdleTime(0)
	db.SetConnMaxLifetime(0)

	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

// ensureDir creates the parent directory for the given path if it does not exist.
func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create db directory %q: %w", dir, err)
	}
	return nil
}

// buildDSN constructs a SQLite DSN with recommended options for the given file path.
func buildDSN(path string) string {
	return fmt.Sprintf("file:%s?_busy_timeout=%d&_journal_mode=WAL&_foreign_keys=ON", url.PathEscape(path), int((5 * time.Second).Milliseconds()))
}

// migrate applies the schema to the database if it is not already present.
func migrate(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}
